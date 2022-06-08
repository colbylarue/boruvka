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
	"fmt"
)

func main() {

	g := new(graph.CGraph)

	//Adding nodes
	g.AddNode() //first node, 0=A
	g.AddNode() //node 1=B
	g.AddNode() //node 2=C
	g.AddNode() //node 3=D
	g.AddNode() //node 4=E
	g.AddNode() //node 5=F
	g.AddNode() //node 6=G
	//fmt.Println("\tOnly nodes, no edges yet:")
	//fmt.Println("nodes:", g.Nodes(), "\nedges:", g.EdgesAll())

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

	fmt.Println("\tNodes and edges:")
	fmt.Println("nodes:", g.Nodes(), "\nedges:", g.EdgesAll())
	fmt.Println("Nr. of one-way edges |E| = ", len(g.EdgesAll()))

	fmt.Println("\tEdges from each node:")
	for _, id := range g.Nodes() {
		fmt.Println("node ", id, "-->", g.EdgesFromNode(id))
	}

	//Testing the method Neighbors
	fmt.Println("\tNeighbors of each node (based on components, not plain edges!):")
	for _, id := range g.Nodes() {
		fmt.Println("Neighbors of ", id, ": ", g.Neighbors(id))
	}

	//Testing the methods NodeMinEdgeSet and NodeMinEdgeGet
	fmt.Println("\tMin edge for each node:")
	for _, id := range g.Nodes() {
		g.NodeMinEdgeSet(id)
		edge := g.NodeMinEdgeGet(id)
		fmt.Println("node ", id, "-->", edge)
		//#######To do: If no min edge was found, this means isolated component
		//#######-----REMOVE FROM THE GRAPH-----
	}

	//Testing the method EdgeContract
	fmt.Println("\tEdge contraction:")
	for _, id := range g.Nodes() {
		edge := g.NodeMinEdgeGet(id)
		if edge[0] == -1 {
			fmt.Println("---No minimum edge!")
		} else {
			g.EdgeContract(edge[0], edge[1])
		}
	}

	//Testing the tree (map)
	fmt.Println("\tTree edges:")
	fmt.Println(graph.Tree)

}
