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
	"boruvka/graph"
	"boruvka/satellite"
	"bufio"
	"fmt"
	"log"
	"os"
)

func parser() []satellite.SimpleSatellite {

	var satlist = []satellite.SimpleSatellite{}

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
	var sat satellite.SimpleSatellite
	for scanner.Scan() {
		s := scanner.Text()
		if linecounter == 0 { // this line is the name of the satellite.
			sat = satellite.SimpleSatellite{Name: s} // TODO: remove whitespace from name first
		} else if linecounter == 1 { // this line is Line 1 of Orbit Info
			sat = satellite.SimpleSatellite{Name: sat.Name, Ole1: s}
		} else if linecounter == 2 { // this line is Line 2 of Orbit info && also final line of data, ready to append
			sat = satellite.SimpleSatellite{Name: sat.Name, Ole1: sat.Ole1, Ole2: s}
			satlist = append(satlist, sat)
			//satG.AddNode() //need to pass name in here
			linecounter = -1
			sat = satellite.SimpleSatellite{}
		}
		linecounter++
	}

	// init all satellites
	for n := range satlist {
		satellite.InitSat(&satlist[n])
	}
	return satlist
}

func main() {

	//########## Initialize graph ######################
	g, gdot := graph.GraphBuilderCsv("data/graph02_12_nodes_no_BOM.csv")
	//generate dot file
	file, err := os.Create("graph.dot")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	file.WriteString(gdot.String())

	Satellites := parser()

	//fmt.Println(Satellites)

	//generate dot file
	f, err := os.Create("graph.dot")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	f.WriteString(gdot.String())

	fmt.Println(Satellites)

	g.Snapshot()

	for g.GetNrNodes() > 1 {
		fmt.Println("#######################################################")
		fmt.Println("##################### MAIN LOOP #######################")
		fmt.Println("#######################################################")
		fmt.Println(g.GetNrNodes(), "nodes in the graph")
		//Calculating the minimum edges for each node in the graph
		fmt.Println("\tMin edge for each node:")
		for _, id := range g.Nodes() {
			if id[1] < 0 {
				fmt.Println("Node already contracted:", id[0])
			} else {
				g.NodeMinEdgeSet(id[1])
				edge := g.NodeMinEdgeGet(id[1])
				fmt.Println("node ", id[1], "--> minEdge:", edge)
			}
			//#######To do: If no min edge was found, this means isolated component
			//#######-----REMOVE FROM THE GRAPH-----
		}

		//Edge Contraction is a multi-step process. It starts with a first pass
		//that adds all minEdges to the Tree and fills up ContractionPairsSlice
		g.BuildContractionPairsSlice()
		fmt.Println("ContractionPairsSlice:", graph.ContractionPairsSlice)
		//Testing the tree (map)
		fmt.Println("\tTree edges:")
		fmt.Println(graph.Tree)

		//This is a process equivalent to Pointer-jumping. We create a slice
		//of leaves (terminal nodes) in leafSlice, and contract those
		for graph.LenContractionPairsSlice() > 0 {
			//leafSlice has a 3rd position that remembers the pair index from
			//ContractionsPairsSlice, to allow fast "deletion"
			leafSlice := make([][3]int, 0)
			//Use ContractionPairs to find the set (slice) of "leaf" pairs for
			//contraction: pairs with one node (or both) appearing only once in
			//ContractionPairsSlice. Unlike ContractionPairsSlice, leafSlice is
			//ordered: The first node will be contracted in the second.
			for i, v := range graph.ContractionPairsSlice {
				fmt.Println("i = ", i, "; v = ", v)
				//#### Optimization: count v[0] and v[1] in the same loop, then
				//examine the counters and decide.
				if graph.OnlyOnceInSlice(v[0], graph.ContractionPairsSlice) {
					leafSlice = append(leafSlice, [3]int{v[0], v[1], i})
				} else if graph.OnlyOnceInSlice(v[1], graph.ContractionPairsSlice) {
					leafSlice = append(leafSlice, [3]int{v[1], v[0], i})
				} //else do nothing - if they both appear more than once, it's not a leaf edge
			}
			fmt.Println("\n############### leafSlice ################\n", leafSlice)
			//Perform a round of leaf contractions according to leafSlice
			if len(leafSlice) > 0 {
				for _, v := range leafSlice {
					g.EdgeContract(v[0], v[1])
					fmt.Println("nodes:", g.Nodes(), "\nedges:", g.EdgesAllMap())
					//Delete the pair from ContractionPairs
					graph.ContractionPairsSlice[v[2]] = [2]int{-1, -1}
				}
			}
		}
	}

}
