package satellite

import (
	"boruvka/graph"
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"os"
	"sort"
	"strings"
)

// satellite data generated from https://www.celestrak.com/NORAD/elements/table.php?GROUP=active&FORMAT=tle
// TODO: need to check copyright or PR

var counter int = 0

type Pair struct {
	Id     int
	Weight int
}

type PairList []Pair

func (p PairList) Len() int           { return len(p) }
func (p PairList) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
func (p PairList) Less(i, j int) bool { return p[i].Weight < p[j].Weight }

// type SimpleSatellite struct
// SimpleSatellite is intended to abstract away all of the Satellite orbital calculations.
// Contains Id, Name, Line 1, Line 2, , MaxEA, LLA, Map of "connected" Sats ids to weights based on distance
// https://en.wikipedia.org/wiki/Two-line_element_set
type SimpleSatellite struct {
	Id            int        `json:"id"`
	Name          string     `json:"name"`
	Tle1          string     `json:"-"`
	Tle2          string     `json:"-"`
	Lla           LatLongAlt `json:"position"`
	MaxEA         float64    `json:"-"`
	PerceivedSats PairList   `json:"perception"`
	MSTneighbors  PairList   `json:"mst"`
}

//lint:ignore U1000 Ignore unused function
func (s *SimpleSatellite) ToString() [3]string {
	return [3]string{s.Name, s.Tle1, s.Tle2}
}

// convert to json https://blog.logrocket.com/using-json-go-guide/
func (s *SimpleSatellite) ToJson() string {
	bytes, _ := json.MarshalIndent(s, "", "\t")
	//fmt.Println(string(bytes))
	return string(bytes)
}

// func initSat pulls the TLE data to generate a LLA position and sets the LLA variable.
// Must be called to populate data
// http://celestrak.org/columns/v02n03/
func InitSat(s *SimpleSatellite) {
	s.Id = counter
	counter = counter + 1
	s.PerceivedSats = make(PairList, 0)
	temp_sat := TLEToSat(s.Tle1, s.Tle2, GravityWGS84)
	pos, _ := Propagate(temp_sat, 2022, 6, 1, 0, 0, 0) //units are km
	s.Lla = ECIToLLA(pos, GSTimeFromDate(2022, 1, 1, 0, 0, 0))

	s.MaxEA = CalculateMaxEA(s)
}

func CalculateMaxEA(s *SimpleSatellite) float64 {
	// equitorial radius (meters)
	RADIUS_EARTH_METERS := 6378137.0
	groundRadius := RADIUS_EARTH_METERS
	maxAngleAcos := groundRadius / (groundRadius + s.Lla.Altitude)
	maxEA := math.Acos(maxAngleAcos) * RAD2DEG
	return maxEA
}

//latlons in radians
func CalculateEA(lat1, lon1, lat2, lon2 float64) float64 {
	ea := (math.Cos(lat1) * math.Cos(lat2) * math.Cos(lon2-lon1)) + math.Sin(lat1)*math.Sin(lat2)

	if ea > 1.0 {
		ea = 1.0
	}
	if ea < -1.0 {
		ea = -1.0
	}
	return math.Acos(ea)
}

// ea is in radians
func CalculateElevationAngle(earthAngle float64, altitude float64) float64 {
	ea := math.Sin(earthAngle) /
		math.Sqrt(1+math.Pow(RADIUS_EARTH_METERS/(altitude+RADIUS_EARTH_METERS), 2)-
			2*(RADIUS_EARTH_METERS/(altitude+RADIUS_EARTH_METERS)*math.Cos(earthAngle)))
	if ea > 1.0 {
		ea = 1.0
	}
	if ea < -1.0 {
		ea = -1.0
	}
	var ground_el float64
	if math.Cos(earthAngle)*(altitude+RADIUS_EARTH_METERS) < RADIUS_EARTH_METERS {
		ground_el = -math.Acos(ea)
	} else {
		ground_el = math.Acos(ea)
	}
	return ground_el
}

//This method has to be called on every satellite against every satellite to determine who can talk
func (s *SimpleSatellite) Discovery(list_all_sats []SimpleSatellite) {

	for i := 0; i < len(list_all_sats); i++ {
		//if i is self continue
		if list_all_sats[i].Name == s.Name {
			continue
		}
		var d_m = CalculateDistanceFromTwoLLA(list_all_sats[i].Lla, s.Lla)
		var eo = CalculateEarthOcclusion(list_all_sats[i].Lla, s.Lla)
		fmt.Println(eo)
		// satellites can really only communicate out a certain distance
		if math.Round(math.Abs(d_m)) >= 10000 {
			continue
		}

		// is earth blocking view? if d_m == -1 then the earth is not occluding the region
		if d_m < 0 {

			s.PerceivedSats = append(s.PerceivedSats, Pair{i, int(math.Round(d_m))})
		}
	}

	sort.Sort(s.PerceivedSats)
}

// func Parser reads a text file line by line to create and initialize simple satellites
// Line 0 : Name
// Line 1 : Line 1 orbitals
// Line 2 : Line 2 orbitals
func Parser(filepath string) []SimpleSatellite {

	var satlist = []SimpleSatellite{}

	file, err := os.Open(filepath)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	//graph to hold sats
	//satG := new(graph.CGraph)

	// TODO: Make this better, add scanner error checking, corrupt data read and filter, etc
	scanner := bufio.NewScanner(file)
	linecounter := 0
	var sat SimpleSatellite
	for scanner.Scan() {
		s := scanner.Text()
		if linecounter == 0 { // this line is the name of the satellite.
			sat = SimpleSatellite{Name: strings.TrimSpace(s)} // TODO: remove whitespace from name first
		} else if linecounter == 1 { // this line is Line 1 of Orbit Info
			sat = SimpleSatellite{Name: sat.Name, Tle1: strings.TrimSpace(s)}
		} else if linecounter == 2 { // this line is Line 2 of Orbit info && also final line of data, ready to append
			sat = SimpleSatellite{Name: sat.Name, Tle1: sat.Tle1, Tle2: strings.TrimSpace(s)}
			satlist = append(satlist, sat)
			//satG.AddNode() //need to pass name in here
			linecounter = -1
			sat = SimpleSatellite{}
		}
		linecounter++
	}

	// init all satellites
	for n := range satlist {
		InitSat(&satlist[n])
	}

	// perform discovery
	for n := range satlist {
		satlist[n].Discovery(satlist)
	}
	return satlist
}

func GenerateCzml(list_all_sats []SimpleSatellite) {

	generatedJSON := "{"
	generatedJSON += `"entities"`
	generatedJSON += ":["
	for i := 0; i < len(list_all_sats); i++ {
		generatedJSON += list_all_sats[i].ToJson()
		if i != len(list_all_sats)-1 {
			generatedJSON += "," // Don't add if it is the last one
		}
	}
	generatedJSON += "]}"

	file, err := os.Create("out/data.json")
	if err != nil {
		fmt.Println(err)
	}
	defer file.Close()

	file.WriteString(strings.Join(strings.Fields(generatedJSON), ""))
}

func ConvertToCGraph(list_all_sats []SimpleSatellite) (g *graph.CGraph) {
	//for now assume it is sorted but better to double check
	// assume id is ascending starting from 0
	// assume id doesn't skip a number
	g = new(graph.CGraph)
	// iterate the list and build graph
	for sat := 0; sat < len(list_all_sats); sat++ {
		//add node from satellite
		sat_id := g.AddNode()
		if sat_id != sat {
			fmt.Println("Sanity Check: Something went wrong mismatched Ids")
		}
	}
	fmt.Println("Status: Nodes added")
	fmt.Println(g.GetNrNodes())
	for sat := 0; sat < len(list_all_sats); sat++ {
		this_sat := list_all_sats[sat]
		//iterate neighbors to add edges and weights
		for key, element := range list_all_sats[sat].PerceivedSats {
			fmt.Println("Key:", key, "=>", "Element:", element)
			other_sat := key
			// Need to check before adding edge to make sure it doesn't alread exist
			// I think it will overwrite it anyway
			g.AddEdgeBoth(this_sat.Id, list_all_sats[other_sat].Id, element.Weight)
		}
	}
	fmt.Println("Status: Edges added")
	return g
}

func GenerateMST(list_all_sats []SimpleSatellite) (g *graph.CGraph) {
	g = ConvertToCGraph(list_all_sats)
	g.BuildMSTBoruvka()
	fmt.Println("FINAL STEP")
	var mstGraph = graph.PrintMSTSorted()
	for i := 0; i < len(mstGraph); i++ {
		//[[0 1 45] [0 2 48]]
		list_all_sats[mstGraph[i][0]].MSTneighbors = append(list_all_sats[mstGraph[i][0]].MSTneighbors, Pair{mstGraph[i][1], mstGraph[i][2]})
	}
	return g
}
