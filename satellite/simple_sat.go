package satellite

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"os"
	"strings"
)

// satellite data generated from https://www.celestrak.com/NORAD/elements/table.php?GROUP=active&FORMAT=tle
// TODO: need to check copyright or PR

// type SimpleSatellite struct
// SimpleSatellite is intended to abstract away all of the Satellite orbital calculations.
// Contains Name, Line 1, Line 2, LLA, Array of "connected" Sats (sats in view)
// https://en.wikipedia.org/wiki/Two-line_element_set
type SimpleSatellite struct {
	Name          string     `json:"name"`
	Tle1          string     `json:"-"`
	Tle2          string     `json:"-"`
	Lla           LatLongAlt `json:"position"`
	MaxEA         float64    `json:"-"`
	PerceivedSats []string   `json:"-"`
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
		satVisible := false
		//if i is self continue
		if list_all_sats[i].Name == s.Name {
			continue
		}
		//. check earth angle
		lat1Rad := DEG2RAD * s.Lla.Latitude
		lat2Rad := DEG2RAD * list_all_sats[i].Lla.Latitude
		lon1Rad := DEG2RAD * s.Lla.Longitude
		lon2Rad := DEG2RAD * list_all_sats[i].Lla.Longitude
		// TODO: refactor to utility function findAngle() and findDistanceTwoPoints()
		earthAngleSatellite := CalculateEA(lat1Rad, lon1Rad, lat2Rad, lon2Rad) * RAD2DEG
		//elevationAngle := CalculateElevationAngle(earthAngleSatellite, s.Lla.Altitude)
		theta := float64(lon1Rad - lon2Rad)
		radtheta := float64(math.Pi * theta / 180)

		dist := math.Sin(lat1Rad)*math.Sin(lat2Rad) + math.Cos(lat1Rad)*math.Cos(lat2Rad)*math.Cos(radtheta)
		if dist > 1 {
			dist = 1.0
		}
		if dist < 1 {
			dist = -1.0
		}

		dist = math.Acos(dist)
		dist = dist * 180 / math.Pi
		//lint:ignore SA4006 Ignore unused value
		dist = dist * 60 * 1.1515
		//fmt.Println(angleToSatellite)
		if earthAngleSatellite > s.MaxEA {
			// check distance
			satVisible = true
		}
		// if sat is visible add to satlist
		if satVisible {
			s.PerceivedSats = append(s.PerceivedSats, list_all_sats[i].Name)
		}
	}
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
		generatedJSON += ","
	}
	generatedJSON += "]}"

	file, err := os.Create("out/data.json")
	if err != nil {
		fmt.Println(err)
	}
	defer file.Close()

	file.WriteString(strings.Join(strings.Fields(generatedJSON), ""))
}
