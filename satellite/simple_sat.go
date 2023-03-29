package satellite

import (
	"boruvka/graph"
	"bufio"
	"fmt"
	"log"
	"math"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

// satellite data generated from https://www.celestrak.com/NORAD/elements/table.php?GROUP=active&FORMAT=tle
// TODO: need to check copyright or PR

var counter int = 0

type Pair struct {
	Id     int
	Weight int
}

type PairList []Pair

func (p PairList) Len() int              { return len(p) }
func (p PairList) Swap(i, j int)         { p[i], p[j] = p[j], p[i] }
func (p PairList) Less(i, j int) bool    { return p[i].Weight < p[j].Weight }
func (p PairList) Greater(i, j int) bool { return p[i].Weight > p[j].Weight }

// type SimpleSatellite struct
// SimpleSatellite is intended to abstract away all of the Satellite orbital calculations.
// Contains Id, Name, Line 1, Line 2, , MaxEA, LLA, Map of "connected" Sats ids to weights based on distance
// https://en.wikipedia.org/wiki/Two-line_element_set
type SimpleSatellite struct {
	Id            int
	Name          string
	Tle1          string
	Tle2          string
	PosECI        Vector3
	Lla           LatLongAlt
	PerceivedSats PairList
	MSTneighbors  PairList
}

func (s *SimpleSatellite) ToString() string {
	sat := `"id":`
	sat += strconv.Itoa(s.Id)
	sat += "," + `"name":"`
	sat += s.Name
	sat += `"` + "," + `"pos":{`
	sat += fmt.Sprintf(`"Lat": %g`, s.Lla.Latitude)
	sat += ","
	sat += fmt.Sprintf(`"Lon": %g`, s.Lla.Longitude)
	sat += ","
	sat += fmt.Sprintf(`"Alt": %g`, s.Lla.Altitude)
	sat += "}"
	return sat
}

func (s *SimpleSatellite) GetPerception() string {
	perception := `"percept": [`
	for i := 0; i < len(s.PerceivedSats); i++ {
		perception += `{"Id": `
		perception += strconv.Itoa(s.PerceivedSats[i].Id)
		perception += `,"Wt": `
		perception += strconv.Itoa(s.PerceivedSats[i].Weight)
		perception += "}"
		if i != len(s.PerceivedSats)-1 {
			perception += ","
		}
	}
	perception += "]"
	return perception
}

func (s *SimpleSatellite) GetMSTneighbors() string {
	mst := `"mst": [`
	for i := 0; i < len(s.MSTneighbors); i++ {
		mst += `{"Id": `
		mst += strconv.Itoa(s.MSTneighbors[i].Id)
		mst += `,"Wt": `
		mst += strconv.Itoa(s.MSTneighbors[i].Weight)
		mst += "}"
		if i != len(s.MSTneighbors)-1 {
			mst += ","
		}
	}
	mst += "]"
	return mst
}

// func initSat pulls the TLE data to generate a LLA position and sets the LLA variable.
// Must be called to populate data
// http://celestrak.org/columns/v02n03/
func InitSat(s *SimpleSatellite) bool {
	s.PerceivedSats = make(PairList, 0)
	temp_sat := TLEToSat(s.Tle1, s.Tle2, GravityWGS84)
	pos, _ := Propagate(temp_sat, 2023, 3, 19, 0, 0, 0) //units are km
	s.PosECI.X = pos.X
	s.PosECI.Y = pos.Y
	s.PosECI.Z = pos.Z

	// Test if the orbit degraded and remove the sat from the list if it did
	// I set the ECI position to (0.0, 0.0, 0.0), i.e. the center of the earth
	// if the orbit has sufficiently degraded to not be a valid prediction
	// or the satellite has burned up in the atmosphere.
	if s.PosECI.X == 0.0 && s.PosECI.Y == 0.0 && s.PosECI.Z == 0.0 {
		return false
	}
	s.Id = counter
	counter = counter + 1
	s.Lla = ECIToLLA(pos, GSTimeFromDate(2023, 2, 19, 0, 0, 0))

	return true
}

//This method has to be called on every satellite against every satellite to determine who can talk
func (s *SimpleSatellite) Discovery(list_all_sats []SimpleSatellite) {

	for i := 0; i < len(list_all_sats); i++ {
		//if i is self continue
		if list_all_sats[i].Name == s.Name {
			continue
		}
		var d_km = CalculateDistanceFromTwoLLA(list_all_sats[i].Lla, s.Lla)
		// satellites can really only communicate out a certain distance
		if math.Round(math.Abs(d_km)) <= 1500 {
			s.PerceivedSats = append(s.PerceivedSats, Pair{list_all_sats[i].Id, int(math.Round(d_km))})
		}

		// is earth blocking view? if eo is true then the earth is occluding the region
		//var eo = CalculateEarthOcclusion(list_all_sats[i].PosECI, s.PosECI)
		////fmt.Println(eo)
		//if !eo {
		//	s.PerceivedSats = append(s.PerceivedSats, Pair{list_all_sats[i].Id, int(math.Round(d_km))})
		//}
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
		isDuplicate := false
		s := scanner.Text()
		if linecounter == 0 { // this line is the name of the satellite.
			sat = SimpleSatellite{Name: strings.TrimSpace(s)} // TODO: remove whitespace from name first
		} else if linecounter == 1 { // this line is Line 1 of Orbit Info
			sat = SimpleSatellite{Name: sat.Name, Tle1: strings.TrimSpace(s)}
		} else if linecounter == 2 { // this line is Line 2 of Orbit info && also final line of data, ready to append
			sat = SimpleSatellite{Name: sat.Name, Tle1: sat.Tle1, Tle2: strings.TrimSpace(s)}
			// Initialize satellite, if valid add to list
			for n := range satlist {
				//check if it is a duplicate
				if sat.Name == satlist[n].Name {
					isDuplicate = true
				}
			}
			if !isDuplicate && InitSat(&sat) {
				satlist = append(satlist, sat)
			}
			linecounter = -1
			sat = SimpleSatellite{}
		}
		linecounter++
	}
	var m sync.Mutex

	// perform discovery
	for n := range satlist {
		m.Lock()
		satlist[n].Discovery(satlist)
		m.Unlock()
	}
	numedges := 0
	for e := range satlist {
		p := len(satlist[e].PerceivedSats)
		numedges += p
	}
	fmt.Println("number of edges = ", numedges)
	return satlist
}

func GenerateCzmlPositions(list_all_sats []SimpleSatellite) {

	generatedJSON := "{"
	generatedJSON += `"entities"`
	generatedJSON += ":["
	for i := 0; i < len(list_all_sats); i++ {
		generatedJSON += "{"
		generatedJSON += list_all_sats[i].ToString()
		generatedJSON += "}"
		if i != len(list_all_sats)-1 {
			generatedJSON += "," // Don't add if it is the last one
		}
	}
	generatedJSON += "]}"

	file, err := os.Create("out/data_positions.json")
	if err != nil {
		fmt.Println(err)
	}
	defer file.Close()

	file.WriteString(strings.Join(strings.Fields(generatedJSON), ""))
}

func GenerateCzmlPerception(list_all_sats []SimpleSatellite) {

	generatedJSON := "{"
	generatedJSON += `"entities"`
	generatedJSON += ":["
	for i := 0; i < len(list_all_sats); i++ {
		generatedJSON += "{"
		generatedJSON += list_all_sats[i].ToString()
		generatedJSON += ","
		generatedJSON += list_all_sats[i].GetPerception()
		generatedJSON += "}"
		if i != len(list_all_sats)-1 {
			generatedJSON += "," // Don't add if it is the last one
		}
	}
	generatedJSON += "]}"

	file, err := os.Create("out/data_perception.json")
	if err != nil {
		fmt.Println(err)
	}
	defer file.Close()

	file.WriteString(strings.Join(strings.Fields(generatedJSON), ""))
}

func GenerateCzmlMst(list_all_sats []SimpleSatellite) {

	generatedJSON := "{"
	generatedJSON += `"entities"`
	generatedJSON += ":["
	for i := 0; i < len(list_all_sats); i++ {
		generatedJSON += "{"
		generatedJSON += list_all_sats[i].ToString()
		generatedJSON += ","
		generatedJSON += list_all_sats[i].GetMSTneighbors()
		generatedJSON += "}"
		if i != len(list_all_sats)-1 {
			generatedJSON += "," // Don't add if it is the last one
		}
	}
	generatedJSON += "]}"

	file, err := os.Create("out/data_mst.json")
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
		for _, element := range list_all_sats[sat].PerceivedSats {
			//fmt.Println("Key:", key, "=>", "Element:", element)
			other_sat := element.Id
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
	graph.BuildDotFromCGraph(g, "test.dot")

	fmt.Println("startingBoruvka")
	start := time.Now()
	g.BuildMSTBoruvka_Parallel(4)
	elapsed := time.Since(start)
	fmt.Println("==elapsed time==", elapsed)
	fmt.Println("FINAL STEP")
	var mstGraph = graph.PrintMSTSorted()
	for i := 0; i < len(mstGraph); i++ {
		//[[0 1 45] [0 2 48]]
		list_all_sats[mstGraph[i][0]].MSTneighbors = append(list_all_sats[mstGraph[i][0]].MSTneighbors, Pair{mstGraph[i][1], mstGraph[i][2]})
	}
	return g
}
