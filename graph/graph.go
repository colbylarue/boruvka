/*
Package graph provides a Graph data structure (directed & weighted)
Initial library was this one from github:
//https://github.com/dorin131/go-data-structures/blob/master/graph/graph.go
*/
package graph

import "fmt"

var Tree = make(map[[2]int][3]int) //Use map to simplify avoiding duplicates

type CGraph struct { //Component Graph
	nodes []*CGraphNode
}

type CGraphNode struct {
	id    int
	edges map[[2]int][3]int //key is array of 2 components: source and dest
	//value is array of 2 original nodes (source, dest) plus weight
	minEdge [5]int //holds min edge for Boruvka alg.
}

// AddNode : adds a new node to the Graph
func (g *CGraph) AddNode() (id int) {
	id = len(g.nodes)
	g.nodes = append(g.nodes, &CGraphNode{
		id:      id,
		edges:   make(map[[2]int][3]int),
		minEdge: [5]int{-1, -1, -1, -1, -1},
	})
	return
}

//AddEdgeBoth : adds edges in both directions, with same weight
//Nodes are inserted in sorted order for faster comparison later
func (g *CGraph) AddEdgeBoth(n1, n2 int, w int) {
	if n1 < n2 {
		g.nodes[n1].edges[[2]int{n1, n2}] = [3]int{n1, n2, w}
		g.nodes[n2].edges[[2]int{n1, n2}] = [3]int{n1, n2, w}
	} else { //opposite order, sorted
		g.nodes[n1].edges[[2]int{n2, n1}] = [3]int{n2, n1, w}
		g.nodes[n2].edges[[2]int{n2, n1}] = [3]int{n2, n1, w}
	}
}

// Neighbors : returns a slice of node IDs that are linked to this node
// Unlike the prev. version, a map is used internally, in order to
//avoid duplicates. The value in the map (-1 below) is irrelevant.
//Uses the key, i.e. the COMPONENTS (not the nodes!)
func (g *CGraph) Neighbors(id int) []int {
	neighbors := make(map[int]int) //an empty map
	for _, node := range g.nodes {
		for edge := range node.edges { //edge is the key (no value)
			if edge[0] == id {
				neighbors[edge[1]] = -1 //insert the other node
			} else if edge[1] == id {
				neighbors[edge[0]] = -1 //insert the other node
			}
		}
	}
	nslice := make([]int, 0) //convert map keys to slice
	for k := range neighbors {
		nslice = append(nslice, k)
	}
	return nslice
}

//Returns a slice of node IDs
func (g *CGraph) Nodes() []int {
	nodes := make([]int, len(g.nodes))
	for i := range g.nodes {
		nodes[i] = i
	}
	return nodes
}

//Returns a slice representing the min edge for the node
func (g *CGraph) NodeMinEdgeGet(id int) [5]int {
	return g.nodes[id].minEdge
}

func (g *CGraph) NodeMinEdgeSet(id int) {
	minE := [5]int{-1, -1, -1, -1, int(1e9)}
	for k, v := range g.nodes[id].edges {
		if v[2] < minE[4] { //found new min
			minE[0] = k[0]
			minE[1] = k[1]
			minE[2] = v[0]
			minE[3] = v[1]
			minE[4] = v[2]
		}
	}
	g.nodes[id].minEdge = minE //OK to copy arrays in Go!
}

//Returns a slice of all edges in the graph, with weights
func (g *CGraph) EdgesAll() [][5]int {
	edges := make([][5]int, 0, len(g.nodes))
	for id := range g.nodes {
		for k, v := range g.nodes[id].edges {
			edges = append(edges, [5]int{k[0], k[1], v[0], v[1], v[2]})
			//Two comps from key, two nodes and the weight from the value
		}
	}
	return edges
}

//Returns a list of all edges from given node, with weights
func (g *CGraph) EdgesFromNode(id int) [][5]int {
	edges := make([][5]int, 0, len(g.nodes[id].edges))
	for k, v := range g.nodes[id].edges {
		edges = append(edges, [5]int{k[0], k[1], v[0], v[1], v[2]})
		//Two comps from key, two nodes and the weight from the value
	}
	return edges
}

//Contracts the components c1 and c2 into a comp. having as id the
//smaller of the two
func (g *CGraph) EdgeContract(c1, c2 int) {
	if c1 > c2 { //nodes and components should always be sorted!
		fmt.Println("Component sorting Error!!")
	}
	fmt.Println("EdgeContract:", c1, c2)
	//adding the edge to the tree (only once)
	Tree[[2]int{c1, c2}] = g.nodes[c1].edges[[2]int{c1, c2}]
	//removing the edge from both components
	delete(g.nodes[c1].edges, [2]int{c1, c2})
	delete(g.nodes[c2].edges, [2]int{c1, c2})
	//invalidating the minEdge for both components (to avoid two-way bridges)
	g.nodes[c1].minEdge = [5]int{-1, -1, -1, -1, -1}
	g.nodes[c2].minEdge = [5]int{-1, -1, -1, -1, -1}

	//#########To do: merge the maps of edges into c1, delete the map of c2
	//for memory efficiency, and "remove" c2 from the array of nodes (set id = -1)
}
