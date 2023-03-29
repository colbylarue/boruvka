/*The function BuildMSTBoruvka_Parallel in this module replaces the old BuildMSTBoruvka.
It takes as argument a nr. of workers (for now either 1, 2, or 4), and finds the
minimum edge for each graph vertex in parallel.
*/
package graph

import (
	"fmt"
	"sort"
	"sync"
	"time"
)

//Holds the tree edges, in the "2-3" format:
//--The key is the [2]int array holding the current components that the edge
//  connects
//--The value is the [3]int array holding the original node numbers connected
// by that edge, and its weight
var Tree = make(map[[2]int][3]int)

var V int = 0 //nr. of vertices will be read from file by Builder function
var ContractionPairsSlice = make([][2]int, 0)
var wg sync.WaitGroup

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

func (n *CGraphNode) GetId() int {
	return n.id
}

func (g *CGraph) GetNode(id int) CGraphNode {
	return *g.nodes[id]
}

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
	return len(g.nodes)
}
func (g *CGraph) GetNrComps() int {
	return g.nrNodes
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
// A map is used internally, in order to avoid duplicates. The value in the
//map (42 below) is irrelevant.
// Uses the key (the "2" part of the "2-3", i.e. COMPONENTS, not original nodes!)
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

func t_decor(f func(g *CGraph)) func(g *CGraph) {
	//wrapper function
	return func(g *CGraph) {
		start := time.Now()
		f(g)
		elapsed := time.Since(start)
		fmt.Println("==elapsed time==", elapsed)
	}
}

//Prints a snapshot of the graph
func (g *CGraph) Snapshot() {
	fmt.Println("\n#### snapshot ############################### \n\tNodes and edges:")
	fmt.Println("nodes:", g.Nodes(), "\nedges:", g.EdgesAllMap())
	fmt.Println("Nr. of undirected edges |E| = ", g.GetNrEdges())
	fmt.Println("\tEdges from each node:")
	for _, id := range g.Nodes() {
		if id[1] >= 0 {
			fmt.Println("node ", id[1], "-->", g.EdgesFromNode(id[1]))
		}
	}
	fmt.Println("\tNeighbors of each node (acc. to components, not plain edges!):")
	for _, id := range g.Nodes() {
		if id[1] >= 0 {
			fmt.Println("Neighbors of ", id[1], ": ", g.Neighbors(id[1]))
		}
	}
	fmt.Println("ContractionsPairsSlice:", ContractionPairsSlice)
	fmt.Println("#############################################")
}

func (g *CGraph) CopyInitialGraph() *CGraph {
	c := new(CGraph)
	for id := 0; id < len(g.nodes); id++ {
		c.AddNode()
		//copy the map of edges for the new node
		for k, v := range g.nodes[id].edges {
			c.nodes[id].edges[k] = v
		}
	}
	return g
}

//Returns a slice representing the min edge for the node
func (g *CGraph) NodeMinEdgeGet(id int) [5]int {
	return g.nodes[id].minEdge
}

//##### For parallel implementation, this function was replaced by the next one
//Change made 2023-03-26: Node/components w/o edges are immediately
//marked as absorbed, with id = -1.
func (g *CGraph) NodeMinEdgeSet(i int) {
	if len(g.nodes[i].edges) == 0 { //Component has no edges
		g.nodes[i].id = -1 //Mark immediately as contracted
		//fmt.Println("This node/component is fully contracted:", i)
	} else {
		minE := [5]int{-1, -1, -1, -1, int(1e9)}
		for k, v := range g.nodes[i].edges {
			if v[2] < minE[4] { //found new min
				minE[0] = k[0]
				minE[1] = k[1]
				minE[2] = v[0]
				minE[3] = v[1]
				minE[4] = v[2]
			} else if v[2] == minE[4] { //implement deterministic tie-break rule
				//This is technically not needed, but, since the order in the set
				//of edges is not guaranteed, it ensures repeatability (for testing
				//and real-life stability).
				//One of the endpoints is the current node, so the sum below
				//says the the other node of the new edge has a smaller id. This
				//avoids spending extra time to sort.
				if k[0]+k[1] < minE[0]+minE[1] {
					minE[0] = k[0]
					minE[1] = k[1]
					minE[2] = v[0]
					minE[3] = v[1]
					minE[4] = v[2]
				}
			}
		}
		g.nodes[i].minEdge = minE //OK to copy arrays in Go!
	}
}

//Similar to NodeMinEdgeSet above, but acts on an entire slice of nodes
//Used by the function BuildMSTBoruvka_Parallel
//Change made 2023-03-26: Node/components w/o edges are immediately
//marked as absorbed, with id = -1.
func (g *CGraph) NodeMinEdgeSetSlice(sliceOfNodes []*CGraphNode) {
	defer wg.Done()
	for i, node := range sliceOfNodes {
		if node.id >= 0 { //non-contracted node
			if len(g.nodes[i].edges) == 0 { //Component has no edges
				g.nodes[i].id = -1 //Mark immediately as contracted
				//fmt.Println("This node/component is fully contracted:", i)
			} else {
				minE := [5]int{-1, -1, -1, -1, int(1e9)}
				for k, v := range g.nodes[i].edges {
					if v[2] < minE[4] { //found new min
						minE[0] = k[0]
						minE[1] = k[1]
						minE[2] = v[0]
						minE[3] = v[1]
						minE[4] = v[2]
					} else if v[2] == minE[4] { //implement deterministic tie-break rule
						//This is technically not needed, but, since the order in the set
						//of edges is not guaranteed, it ensures repeatability (for testing
						//and real-life stability).
						//One of the endpoints is the current node, so the sum below
						//says the the other node of the new edge has a smaller id. This
						//avoids spending extra time to sort.
						if k[0]+k[1] < minE[0]+minE[1] {
							minE[0] = k[0]
							minE[1] = k[1]
							minE[2] = v[0]
							minE[3] = v[1]
							minE[4] = v[2]
						}
					}
				}
				g.nodes[i].minEdge = minE //OK to copy arrays in Go!
			}
		}
	}
}

//Returns the nr. of undirected edges in the graph (no duplicates!)
func (g *CGraph) GetNrEdges() int {
	a := 0
	for i := 0; i < len(g.nodes); i++ {
		a += len(g.nodes[i].edges)
	}
	if a%2 == 0 {
		return a / 2
	} else {
		return -1
	}
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
		if p == pair { //it is legal to compare arrays for equality in Go
			notFound = false
			break
		}
	}
	return notFound
}

//finds out if node p appears only once in the slice of pairs
func OnlyOnceInSlice(p int, sli [][2]int) bool {
	counter := 0
	for _, pair := range sli {
		if p == pair[0] || p == pair[1] {
			counter++
		}
	}
	//fmt.Println("OnlyOnceInSlice: counter =", counter)
	if counter == 1 {
		return true
	} else {
		return false
	}
}

//This function also adds the edges to the MST, so it is OK to delete them later during contraction
func (g *CGraph) BuildContractionPairsSlice() {
	//fmt.Println("\tBuilding ContractionPairsSlice:")
	//Since Go is garbage-collected, there is no memory leak here!
	ContractionPairsSlice = make([][2]int, 0) //Reset this global slice
	for i, n := range g.nodes {
		if n.id >= 0 {
			edge := g.nodes[i].minEdge
			if edge[0] == -1 {
				//fmt.Println("node", i, " has no minimum edge!")
			} else {
				c1, c2 := edge[0], edge[1]
				//Avoiding duplicated edges in Tree and ContractionPairs
				if PairNotInSlice([2]int{c1, c2}, ContractionPairsSlice) {
					//fmt.Println("Adding edge", c1, "-", c2, "to Tree")
					Tree[[2]int{c1, c2}] = g.nodes[c1].edges[[2]int{c1, c2}]
					//fmt.Println("Adding edge", c1, "-", c2, "to ContractionPairsSlice")
					ContractionPairsSlice = append(ContractionPairsSlice, [2]int{c1, c2})
				}
			}
		}
	}
}

//Returns the true length of the slice (ignoring -1 markers)
func LenContractionPairsSlice() int {
	count := 0
	for _, v := range ContractionPairsSlice {
		if v[0] >= 0 {
			count++
		}
	}
	return count
}

//v0 is assimilated into v1
func (g *CGraph) EdgeContract(v0, v1 int) {
	//fmt.Println("\n############## EdgeContract:", v0, "into", v1)
	//Deleting the edge from both components
	sortedv0, sortedv1 := min(v0, v1), max(v0, v1)
	//fmt.Println("Deleting both edges between", sortedv0, "-", sortedv1)
	delete(g.nodes[v0].edges, [2]int{sortedv0, sortedv1})
	delete(g.nodes[v1].edges, [2]int{sortedv0, sortedv1})
	//invalidate the minEdge for the child   ###not really needed - just for testing
	g.nodes[v0].minEdge = [5]int{-1, -1, -1, -1, -1}
	//rename all occurences of v0 (in the map of edges of v0's neighbors) to v1
	//fmt.Println("initial edges from v0:     ", g.EdgesFromNode(v0))
	//#### Idea for later: To reduce writing conflicts, the edges with v0
	//renamed may not be written to v1's map of edges immediately, but stored
	//for now in temporary map
	//c2EdgesMap := make(map[[2]int][3]int)
	for k, v := range g.nodes[v0].edges { //finding all neighbors of v0
		//fmt.Println("\nProcessing neighbor edge: k=", k, "v=", v)
		//First a bit of logic to identify the neighbor:
		neigId := -1
		if k[0] == v0 {
			neigId = k[1]
		} else {
			neigId = k[0]
		}
		//... and a bit of logic to set the neighbor and v1 in sorted order:
		directKey := [2]int{min(neigId, v1), max(neigId, v1)}
		//the edge to the neighbor disappears...
		delete(g.nodes[v0].edges, k)
		delete(g.nodes[neigId].edges, k)
		//... and it is replaced by ...
		//Does the neighbor have a direct edge to v1?
		if directVal, ok := g.nodes[neigId].edges[directKey]; ok {
			//If so, is the direct edge more expensive?
			if directVal[2] > v[2] { //replace direct edge with the smaller one
				g.nodes[neigId].edges[directKey] = v
				g.nodes[v1].edges[directKey] = v
			} //else do nothing, the existing direct edge is the best
		} else { //no direct edge, so add new edge to neighbor's map of edges
			g.nodes[neigId].edges[directKey] = v
			g.nodes[v1].edges[directKey] = v
		}
	}
	//fmt.Println("left over edges should be empty: ", g.EdgesFromNode(v0))
	//Set id = -1 in the CGraphNode structure for v0
	g.nodes[v0].id = -1 //actually deleting would be better, but it's an array
	g.nrNodes--         //decrement the nr. of components left in the graph
}

//To easily compare MSTs generated by different methods (and against pencil-
//and-paper), it is useful to display the final MST as a *sorted* slice
func PrintMSTSorted() [][3]int {
	sli := make([][3]int, 0, len(Tree))
	for _, v := range Tree {
		sli = append(sli, v)
	}
	sort.Slice(sli, func(a, b int) bool {
		if sli[a][0] < sli[b][0] {
			return true
		} else if sli[a][0] == sli[b][0] {
			return sli[a][1] < sli[b][1]
		} else {
			return false
		}
	})
	//fmt.Println("Tree edges SORTED\t:", sli)
	return sli
}

func (g *CGraph) BuildMSTBoruvka() {
	for g.nrNodes > 1 {
		//fmt.Println("\n##################### MAIN LOOP #######################")
		//fmt.Println("#######################################################")
		//fmt.Println(g.nrNodes, "nodes in the graph")
		//Calculating the minimum edge for each node in the graph
		//fmt.Println("\tMin edge for each node:")
		for _, node := range g.Nodes() {
			if node[1] >= 0 {
				g.NodeMinEdgeSet(node[1])
				//edge := g.NodeMinEdgeGet(node[1])
				//fmt.Println("node ", node[1], "--> minEdge:", edge)
			}
			//#######To do: If no min edge was found, this means isolated component
			//#######-----REMOVE FROM THE GRAPH-----
		}

		//Edge Contraction is a multi-step process. It starts with a first pass
		//that adds all minEdges to the Tree and fills up ContractionPairsSlice
		g.BuildContractionPairsSlice()
		//fmt.Println("ContractionPairsSlice:", ContractionPairsSlice)
		//PrintMSTSorted()
		if LenContractionPairsSlice() == 0 { //All connected components have been fully contracted
			//fmt.Println("All connected components have been fully contracted")
			//fmt.Printf("The graph has %d connected components - ### EXITING THE MAIN LOOP ###", g.nrNodes)
			break
		}
		//This is a process equivalent to Pointer-jumping. We create a slice
		//of leaves (terminal nodes) in leafSlice, and contract those
		for LenContractionPairsSlice() > 0 {
			//leafSlice has a 3rd position that remembers the pair index from
			//ContractionsPairsSlice, to allow fast "deletion"
			leafSlice := make([][3]int, 0)
			//Use ContractionPairs to find the set (slice) of "leaf" pairs for
			//contraction: pairs with one node (or both) appearing only once in
			//ContractionPairsSlice. Unlike ContractionPairsSlice, leafSlice is
			//ordered: The first node will be contracted in the second.
			for i, v := range ContractionPairsSlice {
				//fmt.Println("i = ", i, "; v = ", v)
				//#### Optimization: count v[0] and v[1] in the same loop, then
				//examine the counters and decide.
				if OnlyOnceInSlice(v[0], ContractionPairsSlice) {
					leafSlice = append(leafSlice, [3]int{v[0], v[1], i})
				} else if OnlyOnceInSlice(v[1], ContractionPairsSlice) {
					leafSlice = append(leafSlice, [3]int{v[1], v[0], i})
				} //else do nothing - if they both appear more than once, it's not a leaf edge
			}
			//fmt.Println("\n###leafSlice - note that here the nodes of an edge may be unsorted!\n", leafSlice)
			//Perform a round of leaf contractions according to leafSlice
			if len(leafSlice) > 0 {
				for _, v := range leafSlice {
					g.EdgeContract(v[0], v[1])
					//fmt.Println("nodes:", g.Nodes(), "\nedges:", g.EdgesAllMap())
					//Delete the pair from ContractionPairs
					ContractionPairsSlice[v[2]] = [2]int{-1, -1}
				}

			}
		} //end inner for loop (builds leafSlice every time and contracts it)
	} //end outer for loop (builds ContractionPairsSlice every time and contracts it)
}

//The function BuildMSTBoruvka_Parallel replaces the old BuildMSTBoruvka above.
//It takes as argument a nr. of workers (for now either 1, 2, or 4), and finds the
//minimum edge for each graph vertex in parallel, using another new function
//g.NodeMinEdgeSetSlice (declared above).
func (g *CGraph) BuildMSTBoruvka_Parallel(n int) {
	for g.nrNodes > 1 {
		//fmt.Println("\n##################### MAIN LOOP #######################")
		//fmt.Println("#######################################################")
		//fmt.Println(g.nrNodes, "nodes in the graph")
		//Calculating the minimum edge for each node in the graph
		if n == 1 {
			wg.Add(1)
			go g.NodeMinEdgeSetSlice(g.nodes)
			wg.Wait()
		} else if n == 2 {
			wg.Add(2)
			go g.NodeMinEdgeSetSlice(g.nodes[:V/2])
			go g.NodeMinEdgeSetSlice(g.nodes[V/2:])
			wg.Wait()
		} else if n == 4 {
			wg.Add(4)
			go g.NodeMinEdgeSetSlice(g.nodes[:V/4])
			go g.NodeMinEdgeSetSlice(g.nodes[V/4 : V/2])
			go g.NodeMinEdgeSetSlice(g.nodes[V/2 : 3*V/4])
			go g.NodeMinEdgeSetSlice(g.nodes[3*V/4:])
			wg.Wait()
		} else {
			panic("Invalid choice for n!")
		}

		//Edge Contraction is a multi-step process. It starts with a first pass
		//that adds all minEdges to the Tree and fills up ContractionPairsSlice
		g.BuildContractionPairsSlice()
		//fmt.Println("ContractionPairsSlice:", ContractionPairsSlice)
		//PrintMSTSorted()
		if LenContractionPairsSlice() == 0 { //All connected components have been fully contracted
			//fmt.Println("All connected components have been fully contracted")
			//fmt.Printf("The graph has %d connected components - ### EXITING THE MAIN LOOP ###", g.nrNodes)
			break
		}
		//This is a process equivalent to Pointer-jumping. We create a slice
		//of leaves (terminal nodes) in leafSlice, and contract those
		for LenContractionPairsSlice() > 0 {
			//leafSlice has a 3rd position that remembers the pair index from
			//ContractionsPairsSlice, to allow fast "deletion"
			leafSlice := make([][3]int, 0)
			//Use ContractionPairs to find the set (slice) of "leaf" pairs for
			//contraction: pairs with one node (or both) appearing only once in
			//ContractionPairsSlice. Unlike ContractionPairsSlice, leafSlice is
			//ordered: The first node will be contracted in the second.
			for i, v := range ContractionPairsSlice {
				//fmt.Println("i = ", i, "; v = ", v)
				//#### Optimization: count v[0] and v[1] in the same loop, then
				//examine the counters and decide.
				if OnlyOnceInSlice(v[0], ContractionPairsSlice) {
					leafSlice = append(leafSlice, [3]int{v[0], v[1], i})
				} else if OnlyOnceInSlice(v[1], ContractionPairsSlice) {
					leafSlice = append(leafSlice, [3]int{v[1], v[0], i})
				} //else do nothing - if they both appear more than once, it's not a leaf edge
			}
			//fmt.Println("\n###leafSlice - note that here the nodes of an edge may be unsorted!\n", leafSlice)
			//Perform a round of leaf contractions according to leafSlice
			if len(leafSlice) > 0 {
				for _, v := range leafSlice {
					g.EdgeContract(v[0], v[1])
					//fmt.Println("nodes:", g.Nodes(), "\nedges:", g.EdgesAllMap())
					//Delete the pair from ContractionPairs
					ContractionPairsSlice[v[2]] = [2]int{-1, -1}
				}

			}
		} //end inner for loop (builds leafSlice every time and contracts it)
	} //end outer for loop (builds ContractionPairsSlice every time and contracts it)
}
