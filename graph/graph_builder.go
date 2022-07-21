package graph

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/tmc/dot"
)

// TODO: GraphBuilderDot() - generate CGraph from Dot file
//1. read in dot file graph
//2. create dot graph
//3. create CGraph
// Node id must match up if node is -1 ignore

func Contains(nodes []dot.Node, n dot.Node) bool {
	//search existing nodes to see if it already exists
	nodeExists := false
	for j := 0; j < len(nodes); j++ {
		//    Compare: Return 0, if str1 == str2.
		//    Compare: Return 1, if str1 > str2.
		//    Compare: Return -1, if str1 < str2.
		if strings.Compare(nodes[j].Name(), n.Name()) == 0 {
			// This means we have already added it to the node list
			nodeExists = true
			break
		}
	}
	return nodeExists
}

func FindNodeById(nodes []dot.Node, id string) *dot.Node {

	for i := 0; i < len(nodes); i++ {
		//fmt.Printf("FindNodeById(): id= %s Name= %s", id, nodes[i].Name())
		if strings.Compare(nodes[i].Name(), id) == 0 {
			// This means we have found the node
			//fmt.Printf("FindNodeById(): nodes[%d]=%s ", i, nodes[i].Name())
			return &nodes[i]
		}
	}
	// the node didn't match names
	// TODO: Handle error here elegantly
	return dot.NewNode("error")
}

func BuildDotFromCGraph(g *CGraph, outfile string) *dot.Graph {
	gdot := dot.NewGraph("GeneratedGraph")
	gdot.SetType(dot.GRAPH)
	gdot.Set("layout", "dot")
	var nodes []dot.Node

	for i := 0; i < g.nrNodes; i++ {
		if g.nodes[i].id < 0 {
			// -1 this means the node is absorbed into another
			continue
		}
		ndot := dot.NewNode(fmt.Sprint(i))
		nodeExists := Contains(nodes, *ndot)
		if nodeExists {
			// already exists don't add
			continue
		}
		gdot.AddNode(ndot)
		nodes = append(nodes, *ndot)
	}

	var edges [][3]int
	// Now that all the nodes exist we need to add the edges

	for i := 0; i < g.nrNodes; i++ {
		//iterate through each nodes edges
		for key, val := range g.nodes[i].edges {
			//fmt.Println("...Key:", key, "=>", "Element:", val)
			valExists := false
			//check to see if we have already added this edge
			for _, it := range edges {
				if it == val {
					valExists = true
					break
				}
			}
			if !valExists {
				edge := dot.NewEdge(FindNodeById(nodes, strconv.Itoa(key[0])), FindNodeById(nodes, strconv.Itoa(key[1])))
				if edge.Source().Name() == "error" || edge.Destination().Name() == "error" {
					continue
				}
				edge.Set("weight", fmt.Sprint(val[2]))
				edge.Set("label", fmt.Sprint(val[2]))
				gdot.AddEdge(edge)
				edges = append(edges, val)
			}

		}

	}
	//generate dot file
	var filename string
	if outfile == "" {
		filename = "graph_snapshot_0.dot"
	} else {
		filename = outfile
	}

	cwd, _ := os.Getwd()
	err := os.Mkdir("out", os.ModeDir) // you might want different file access, this suffice for this example
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Printf("Created %s at %s\n", "out", cwd)
	}

	for i := 0; ; i++ {
		_, error := os.Stat(filepath.Join(cwd, "out", filename))
		// check if error is "file not exists"
		if os.IsNotExist(error) {
			break
		} else {
			//fmt.Printf("%v file exist\n", filename)
			filename = filename + fmt.Sprint(i) + ".dot"
		}
	}

	path := filepath.Join(cwd, "out", filename)
	newFilePath := filepath.FromSlash(path)
	file, err := os.Create(newFilePath)
	if err != nil {
		fmt.Println(err)
	}
	defer file.Close()

	file.WriteString(gdot.String())
	return gdot
}

func GraphBuilderCsv(csvFilePath string) (c *CGraph, d *dot.Graph) {
	g := new(CGraph)
	gdot := dot.NewGraph("Example Graph")
	gdot.SetType(dot.GRAPH)
	gdot.Set("layout", "dot")
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
	for i := 0; ; i++ {
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
			for i := 0; i < numOfNodes; i++ {
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
