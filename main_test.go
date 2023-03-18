package main

import (
	"boruvka/graph"
	"boruvka/satellite"
	"testing"
)

func TestGraph(t *testing.T) {

	// TODO: add test -> examples in repo below:
	// https://github.com/networkx/networkx/blob/main/networkx/classes/tests/test_graph.py

	t.Run("Build Graph Test", func(t *testing.T) {
		g := new(graph.CGraph)
		expected := 0

		if expected != g.GetNrNodes() {
			t.Errorf("Expected %d; but got %d", expected, g.GetNrNodes())
		}
	})

	t.Run("Build Satellite Graph Test", func(t *testing.T) {
		Satellites := satellite.Parser("satellite/SatDB.txt")
		g := satellite.GenerateMST(Satellites)
		expected := 3520

		if expected != g.GetNrNodes() {
			t.Errorf("Expected %d; but got %d", expected, g.GetNrNodes())
		}
	})
}
