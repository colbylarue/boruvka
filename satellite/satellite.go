//IN PROGRESS

package main

import (
	"boruvka/graph"
	"bufio"
	"log"
	"os"
)

func main() {
	// satellite data generated from https://www.celestrak.com/NORAD/elements/table.php?GROUP=active&FORMAT=tle
	// TODO: need to check copyright or PR

	//type SatelliteGraph struct
	// Note the Satellite Graph will have N^2 connections worst cast (technically line of sight narrows it down) (Look at subgraphs)
	type Satellite struct {
		Name string
		Ole1 string
		Ole2 string
	}
	var satlist = []Satellite{}

	file, err := os.Open("SatDB.txt")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	//graph to hold sats
	satG := new(graph.CGraph)

	// TODO: Make this better, add scanner error checking, corrupt data read and filter, etc
	scanner := bufio.NewScanner(file)
	linecounter := 0
	var sat Satellite
	for scanner.Scan() {
		s := scanner.Text()
		if linecounter == 0 { // this line is the name of the satellite.
			sat = Satellite{Name: s} // TODO: remove whitespace from name first
		} else if linecounter == 1 { // this line is Line 1 of Orbit Info
			sat = Satellite{Name: sat.Name, Ole1: s}
		} else if linecounter == 2 { // this line is Line 2 of Orbit info && also final line of data, ready to append
			sat = Satellite{Name: sat.Name, Ole1: sat.Ole1, Ole2: s}
			satlist = append(satlist, sat)
			satG.AddNode() //need to pass name in here
			linecounter = -1
			sat = Satellite{}
		}
		linecounter++
	}
}
