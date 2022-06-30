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

	"github.com/tmc/dot"
)

// TODO: Convert this to just be a unit test
func buildGraph() *graph.CGraph {
	g := new(graph.CGraph)
	gdot := dot.NewGraph("Example Graph")
	gdot.SetType(dot.GRAPH)
	gdot.Set("layout", "circo")
	//Adding nodes
	g.AddNode() //first node, 0=A
	g.AddNode() //node 1=B
	g.AddNode() //node 2=C
	g.AddNode() //node 3=D
	g.AddNode() //node 4=E
	g.AddNode() //node 5=F
	g.AddNode() //node 6=G
	ndot0 := dot.NewNode("0")
	gdot.AddNode(ndot0)
	ndot1 := dot.NewNode("1")
	gdot.AddNode(ndot1)
	ndot2 := dot.NewNode("2")
	gdot.AddNode(ndot2)
	ndot3 := dot.NewNode("3")
	gdot.AddNode(ndot3)
	ndot4 := dot.NewNode("4")
	gdot.AddNode(ndot4)
	ndot5 := dot.NewNode("5")
	gdot.AddNode(ndot5)
	ndot6 := dot.NewNode("6")
	gdot.AddNode(ndot6)
	//Adding edges
	g.AddEdgeBoth(0, 1, 7)  //AB = 7
	g.AddEdgeBoth(0, 3, 4)  //AD = 4
	g.AddEdgeBoth(1, 2, 11) //BC = 11
	g.AddEdgeBoth(1, 3, 9)  //BD = 9
	g.AddEdgeBoth(1, 4, 10) //BE = 10
	g.AddEdgeBoth(2, 4, 5)  //CE = 5
	g.AddEdgeBoth(3, 4, 15) //DE = 15
	g.AddEdgeBoth(3, 5, 6)  //DF = 6
	g.AddEdgeBoth(4, 5, 12) //EF = 12
	g.AddEdgeBoth(4, 6, 8)  //EG = 8
	g.AddEdgeBoth(5, 6, 13) //FG = 13
	//########## End initialization ###################

	e1 := dot.NewEdge(ndot0, ndot1)
	e1.Set("weight", "7")
	e1.Set("label", "7")
	gdot.AddEdge(e1)
	e2 := dot.NewEdge(ndot0, ndot3)
	e2.Set("weight", "4")
	e2.Set("label", "4")
	gdot.AddEdge(e2)
	e3 := dot.NewEdge(ndot1, ndot2)
	e3.Set("weight", "11")
	e3.Set("label", "11")
	gdot.AddEdge(e3)
	e4 := dot.NewEdge(ndot1, ndot3)
	e4.Set("weight", "9")
	e4.Set("label", "9")
	gdot.AddEdge(e4)
	e5 := dot.NewEdge(ndot1, ndot4)
	e5.Set("weight", "10")
	e5.Set("label", "10")
	gdot.AddEdge(e5)
	e6 := dot.NewEdge(ndot2, ndot4)
	e6.Set("weight", "5")
	e6.Set("label", "5")
	gdot.AddEdge(e6)
	e7 := dot.NewEdge(ndot3, ndot4)
	e7.Set("weight", "15")
	e7.Set("label", "15")
	gdot.AddEdge(e7)
	e8 := dot.NewEdge(ndot3, ndot5)
	e8.Set("weight", "6")
	e8.Set("label", "6")
	gdot.AddEdge(e8)
	e9 := dot.NewEdge(ndot4, ndot5)
	e9.Set("weight", "12")
	e9.Set("label", "12")
	gdot.AddEdge(e9)
	e10 := dot.NewEdge(ndot4, ndot6)
	e10.Set("weight", "8")
	e10.Set("label", "8")
	gdot.AddEdge(e10)
	e11 := dot.NewEdge(ndot5, ndot6)
	e11.Set("weight", "13")
	e11.Set("label", "13")
	gdot.AddEdge(e11)

	//generate dot file
	file, err := os.Create("graph.dot")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	file.WriteString(gdot.String())

	return g
}

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
