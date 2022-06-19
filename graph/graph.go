/*
Package graph provides a Graph data structure (directed & weighted)
Initial library was this one from github:
//https://github.com/dorin131/go-data-structures/blob/master/graph/graph.go
*/
package graph

import (
	"fmt"
)

var Tree = make(map[[2]int][3]int) //Holds the tree edges, in the "2-3" format
var ContractionPairsSlice = make([][2]int, 0)

//var visited = make(map[int]int)

type CGraph struct { //Component Graph
	nrNodes int
	nodes   []*CGraphNode
}

type CGraphNode struct {
	id    int
	edges map[[2]int][3]int //key is array of 2 components: source and dest
	//value is array of 2 original nodes (source, dest) plus weight
	minEdge [5]int //holds min edge for Boruvka alg.
}

/*
func traverse(*CGraph) {
	for node := range adjList {
		if visited[node] == 1 {
			continue
		}
		if !dfs(node, adjList, &visited) {
			return false
		}
	}
}

func dfs(node int, adjList map[int][]int, visited *map[int]int) bool {

	neighbourArr, ok := adjList[node]
	if !ok {
		return true
	}

	if (*visited)[node] == -1 {
		return false
	}

	if (*visited)[node] == 1 {
		return true
	}

	(*visited)[node] = -1

	for _, neighbour := range neighbourArr {
		if !dfs(neighbour, adjList, visited) {
			return false
		}
	}
	(*visited)[node] = 1
	return true
}
*/
/*
func convert() (*cgraph.Graph, error) {
	g := graphviz.New()
	graph, err := g.Graph()
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := graph.Close(); err != nil {
			log.Fatal(err)
		}
		g.Close()
	}()
	n, err := graph.CreateNode("n")
	if err != nil {
		log.Fatal(err)
	}
	m, err := graph.CreateNode("m")
	if err != nil {
		log.Fatal(err)
	}
	e, err := graph.CreateEdge("e", n, m)
	if err != nil {
		log.Fatal(err)
	}
	e.SetLabel("e")
	var buf bytes.Buffer
	if err := g.Render(graph, "dot", &buf); err != nil {
		log.Fatal(err)
	}
	fmt.Println(buf.String())
}*/

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// AddNode : adds a new node to the Graph
func (g *CGraph) AddNode() (id int) {
	id = len(g.nodes)
	g.nodes = append(g.nodes, &CGraphNode{
		id:      id,
		edges:   make(map[[2]int][3]int),
		minEdge: [5]int{-1, -1, -1, -1, -1},
	})
	g.nrNodes = len(g.nodes)
	return
}

func (g *CGraph) GetNrNodes() int {
	return g.nrNodes
}
func (g *CGraph) DecNrNodes() {
	g.nrNodes--
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

	fmt.Println(g)
}

// Neighbors : returns a slice of node IDs that are linked to this node
// Unlike the prev. version, a map is used internally, in order to
//avoid duplicates. The value in the map (-1 below) is irrelevant.
//Uses the key, i.e. the COMPONENTS (not the nodes!)
func (g *CGraph) Neighbors(id int) []int {
	neighbors := make(map[int]int) //an empty map
	for _, node := range g.nodes {
		if node.id >= 0 { //only non-contracted nodes
			for edge := range node.edges { //edge is the key (no value)
				if edge[0] == id {
					neighbors[edge[1]] = 42 //insert the other node
				} else if edge[1] == id {
					neighbors[edge[0]] = 42 //insert the other node
				}
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
func (g *CGraph) Nodes() [][2]int {
	nodes := make([][2]int, len(g.nodes))
	for i := range g.nodes {
		nodes[i] = [2]int{i, g.nodes[i].id}
	}
	return nodes
}

//Prints a snapshot of the graph
func (g *CGraph) Snapshot() {
	fmt.Println("\n#### snapshot ############################### \n\tNodes and edges:")
	fmt.Println("nodes:", g.Nodes(), "\nedges:", g.EdgesAllMap())
	fmt.Println("Nr. of two-way edges |E| = ", len(g.EdgesAllMap()))
	fmt.Println("\tEdges from each node:")
	for _, id := range g.Nodes() {
		if id[1] < 0 {
			fmt.Println("Contracted node:", id[0])
		} else {
			fmt.Println("node ", id[1], "-->", g.EdgesFromNode(id[1]))
		}
	}
	fmt.Println("\tNeighbors of each node (acc. to components, not plain edges!):")
	for _, id := range g.Nodes() {
		if id[1] < 0 {
			fmt.Println("Contracted node:", id[0])
		} else {
			fmt.Println("Neighbors of ", id[1], ": ", g.Neighbors(id[1]))
		}
	}
	fmt.Println("ContractionsPairsSlice:", ContractionPairsSlice)
	fmt.Println("#############################################")
}

//creates & writes graph to a .dot file
//func (g *CGraph) generateDot() {
//
//}

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
		} else if v[2] == minE[4] { //implement tie-break rule
			if k[0]+k[1] < minE[0]+minE[1] { //new edge has a smaller node id
				minE[0] = k[0]
				minE[1] = k[1]
				minE[2] = v[0]
				minE[3] = v[1]
				minE[4] = v[2]
			}
		}
	}
	g.nodes[id].minEdge = minE //OK to copy arrays in Go!
}

//Returns a slice (with duplicates) of all edges in the graph, with weights
func (g *CGraph) EdgesAllSlice() [][5]int {
	edges := make([][5]int, 0, len(g.nodes))
	for id := range g.nodes {
		if g.nodes[id].id >= 0 { //only nodes/components still active
			for k, v := range g.nodes[id].edges {
				edges = append(edges, [5]int{k[0], k[1], v[0], v[1], v[2]})
				//Two comps from key, two nodes and the weight from the value
			}
		}
	}
	return edges
}

//Returns a map (to avoid duplicates) of all edges in the graph, with weights
func (g *CGraph) EdgesAllMap() map[[2]int][3]int {
	edges := make(map[[2]int][3]int)
	for nid := range g.nodes {
		if g.nodes[nid].id >= 0 { //only nodes/components not contracted
			for k, v := range g.nodes[nid].edges {
				edges[k] = v
			}
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

func PairNotInSlice(p [2]int, sli [][2]int) bool {
	notFound := true
	for _, pair := range sli {
		if p == pair {
			notFound = false
			break
		}
	}
	return notFound
}

func ParentNotInSlice(p int, sli [][2]int) bool {
	notFound := true
	for _, pair := range sli {
		if p == pair[0] {
			notFound = false
			break
		}
	}
	return notFound
}

func (g *CGraph) BuildContractionPairsSlice() {
	fmt.Println("\tBuilding ContractionPairsSlice:")
	//Since Go is garbage-collected, there is no memory leak here!
	ContractionPairsSlice = make([][2]int, 0)
	for i, n := range g.nodes {
		if n.id >= 0 {
			edge := g.nodes[i].minEdge
			if edge[0] == -1 {
				fmt.Println("node", i, " has no minimum edge!")
			} else {
				c1, c2 := edge[0], edge[1]
				fmt.Println("Adding edge", c1, "-", c2, "to Tree")
				Tree[[2]int{c1, c2}] = g.nodes[c1].edges[[2]int{c1, c2}]
				//Avoiding duplicated edges in ContractionPairs
				if PairNotInSlice([2]int{c1, c2}, ContractionPairsSlice) {
					fmt.Println("Adding edge", c1, "-", c2, "to ContractionPairsSlice")
					ContractionPairsSlice = append(ContractionPairsSlice, [2]int{c1, c2})
				}
			}
		}
	}
}

//Retuens the true length of the slice (ignoring -1 markers)
func LenContractionPairsSlice() int {
	count := 0
	for _, v := range ContractionPairsSlice {
		if v[0] >= 0 {
			count++
		}
	}
	return count
}

func (g *CGraph) EdgeContract(c1, c2 int) {
	fmt.Println("\n############## EdgeContract:", c2, "into", c1)
	//Deleting the edge from both components
	fmt.Println("Deleting", c1, "-", c2, "and", c2, "-", c1)
	delete(g.nodes[c1].edges, [2]int{c1, c2})
	delete(g.nodes[c2].edges, [2]int{c1, c2})
	//invalidate the minEdge for the child   ###not really needed - just for testing
	g.nodes[c2].minEdge = [5]int{-1, -1, -1, -1, -1}
	//rename all occurences of c2 (in the map of edges of c2) to c1
	fmt.Println("initial edges from c2:     ", g.EdgesFromNode(c2))
	//#### Idea for later: To reduce writing conflicts, the edges with c2
	//renamed may not be written to c1's map of edges immediately, but stored
	//for now in temporary map
	//c2EdgesMap := make(map[[2]int][3]int)
	for k, v := range g.nodes[c2].edges { //all neighbors of c2
		fmt.Println("\nProcessing neighbor edge: k=", k, "v=", v)
		//First a bit of logic to identify the neighbor:
		neigId := -1
		if k[0] == c2 {
			neigId = k[1]
		} else {
			neigId = k[0]
		}
		//... and a bit of logic to set the neighbor and c1 in sorted order:
		directKey := [2]int{min(neigId, c1), max(neigId, c1)}
		//the edge to the neighbor disappears...
		delete(g.nodes[c2].edges, k)
		delete(g.nodes[neigId].edges, k)
		//... and it is replaced by ...
		//Does the neighbor have a direct edge to c1?
		if directVal, ok := g.nodes[neigId].edges[directKey]; ok {
			//If so, is the direct edge more expensive?
			if directVal[2] > v[2] { //replace direct edge with the smaller one
				g.nodes[neigId].edges[directKey] = v
				g.nodes[c1].edges[directKey] = v
			} //else do nothing, the existing direct edge is the best
		} else { //no direct edge, so add new edge to neighbor's map of edges
			g.nodes[neigId].edges[directKey] = v
			g.nodes[c1].edges[directKey] = v
		}
	}
	fmt.Println("left over edges should be empty: ", g.EdgesFromNode(c2))
	//Set id = -1 in the CGraphNode structure for c2
	g.nodes[c2].id = -1
	g.DecNrNodes()
}
