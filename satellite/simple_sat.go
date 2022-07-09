package satellite

import (
	"bufio"
	"log"
	"os"
)

// satellite data generated from https://www.celestrak.com/NORAD/elements/table.php?GROUP=active&FORMAT=tle
// TODO: need to check copyright or PR

// type SimpleSatellite struct
// SimpleSatellite is intended to abstract away all of the Satellite orbital calculations.
// Contains Name, Line 1, Line 2, LLA, Array of "connected" Sats (sats in view)
type SimpleSatellite struct {
	Name string
	Ole1 string
	Ole2 string
	Lla  LatLongAlt
}

//lint:ignore U1000 Ignore unused function
func (s *SimpleSatellite) ToString() [3]string {
	return [3]string{s.Name, s.Ole1, s.Ole2}
}

// func initSat pulls the TLE data to generate a LLA position and sets the LLA variable.
// Must be called to populate data
func InitSat(s *SimpleSatellite) {
	temp_sat := TLEToSat(s.Ole1, s.Ole2, GravityWGS84)
	pos, _ := Propagate(temp_sat, 2022, 6, 1, 0, 0, 0)
	s.Lla = ECIToLLA(pos, GSTimeFromDate(2022, 1, 1, 0, 0, 0))
}

// func Parser reads a text file line by line to create and initialize simple satellites
// Line 0 : Name
// Line 1 : Line 1 orbitals
// Line 2 : Line 2 orbitals
func Parser() []SimpleSatellite {

	var satlist = []SimpleSatellite{}

	//TODO: pass in file
	file, err := os.Open("satellite/SatDB.txt")
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
			sat = SimpleSatellite{Name: s} // TODO: remove whitespace from name first
		} else if linecounter == 1 { // this line is Line 1 of Orbit Info
			sat = SimpleSatellite{Name: sat.Name, Ole1: s}
		} else if linecounter == 2 { // this line is Line 2 of Orbit info && also final line of data, ready to append
			sat = SimpleSatellite{Name: sat.Name, Ole1: sat.Ole1, Ole2: s}
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
	return satlist
}
