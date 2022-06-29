package graph

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"

	"github.com/tmc/dot"
)

//func graphBuilder(*dot.Graph) (*CGraph, *dot.Graph) {
//	g := new(CGraph)
//	gdot := dot.NewGraph("Example Graph")
//	gdot.SetType(dot.GRAPH)
//	gdot.Set("layout", "circo")
//
//	//1. read in dot file graph
//	//2. create dot graph
//	//3. create CGraph
//	// Node id must match up
//
//	return g, gdot
//}

func GraphBuilderCsv(csvFilePath string) (c *CGraph, d *dot.Graph) {
	g := new(CGraph)
	gdot := dot.NewGraph("Example Graph")
	gdot.SetType(dot.GRAPH)
	gdot.Set("layout", "circo")
	// open file
	f, err := os.Open(csvFilePath)
	if err != nil {
		log.Fatal(err)
	}

	// remember to close the file at the end of the program
	defer f.Close()

	// read csv values using csv.Reader
	csvReader := csv.NewReader(f)
	var numOfNodes int
	// create map of fields
	var fieldMap map[string]string
	var nodes []dot.Node
	i := 0
	for {
		rec, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("%+v\n", rec)

		// Parse Data
		if i == 0 {
			mystr := rec[0]
			n, err := strconv.Atoi(mystr)
			fmt.Println(n)
			if err != nil {
				log.Fatal(err)
			}
			numOfNodes = n
			//create nodes in graph id starts at 0
			for i < numOfNodes {
				g.AddNode()

				ndot := dot.NewNode(fmt.Sprint(i))
				nodes = append(nodes, *ndot)
				gdot.AddNode(ndot)
			}
			continue
		} else if i == 1 {
			//for _, field := range rec {
			// TODO: initialize fieldMap
			// each field is a header in the csv
			// since i plan to add attributes to this later
			// i think a map of strings to strings is most flexible
			// for now I have hard coded the values below n1, n2, & w
			//}
			continue
		} else {
			//sanity check here
			if len(rec) >= 2 {
				n1, _ := strconv.Atoi(rec[0])
				n2, _ := strconv.Atoi(rec[1])
				w, _ := strconv.Atoi(rec[2])
				g.AddEdgeBoth(n1, n2, w)

				e1 := dot.NewEdge(&nodes[n1], &nodes[n2])
				e1.Set("weight", fmt.Sprint(w))
				e1.Set("label", fmt.Sprint(w))
				gdot.AddEdge(e1)
			}
		}
		fmt.Println(fieldMap)
		i++
	}

	return g, gdot
}
