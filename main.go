// boruvka02 project main.go
//implementing the Wikipedia Boruvka example:
//https://en.wikipedia.org/wiki/Bor%C5%AFvka%27s_algorithm
//
//Changes to the data structure:
//--Graph is renamed CGraph (component graph); instead of plain nodes,
//the nodes of the CGraph are components (comps), i.e. as merged nodes
//--The edge is still represented as a map element, however:
//	-The key is the array of two elements source-comp, dest-comp
//   (not nodes, but components, since Boruvka merges nodes and then
//	  entire components into larger components!)
//	-The value is another array, with three elements:
//   weight, original-source-node, original-dest-node (original nodes
//   need to be preserved in order to be able to identify the edge when
//   chosen)
//--Accordingly, the nodes are renamed CGraphNodes:
//  -an array of 3 ints was added to hold the minimum edge (empty for now)
//  -the edges incident to the node are still represented as a map,
//   but the key is the full pair source-destination, and in sorted
//   order: source < dest; this will make more efficient the comparison of
//   minimum edges from different components.
//--The set T (tree edges) is implem. as slice of 3-element arrays
//(empty for now).
package main

import (
	"boruvka/satellite"
	"fmt"
	"os"
	"path/filepath"
)

func main() {
	files, err := filepath.Glob(filepath.Join("out", "*"))
	if err != nil {
		fmt.Println(err)
	}
	for _, file := range files {
		err = os.RemoveAll(file)
		if err != nil {
			fmt.Println(err)
		}
	}

	//########## Initialize graph ######################
	//g, gdot := graph.GraphBuilderCsv("data/graph02_12_nodes_no_BOM.csv")
	////generate dot file
	//file, err := os.Create("graph.dot")
	//if err != nil {
	//	log.Fatal(err)
	//}
	//defer file.Close()
	//file.WriteString(gdot.String())

	//Test building Dot from CGraph
	//new_g := graph.BuildDotFromCGraph(g, "")
	//fmt.Println(new_g.String())
	// This method is slow TODO: investigate speedup
	Satellites := satellite.Parser("satellite/StarlinkDB.txt")
	satellite.GenerateMST(Satellites)
	// do this after the MST so the data is populated
	satellite.GenerateCzml(Satellites)

}
