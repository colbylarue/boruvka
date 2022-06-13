package main

import (
	"boruvka/graph"
	"testing"
)

func TestAdd(t *testing.T) {

	// TODO: add test -> examples in repo below:
	// https://github.com/networkx/networkx/blob/main/networkx/classes/tests/test_graph.py

	t.Run("Empty Graph", func(t *testing.T) {
		g := new(graph.CGraph)
		expected := 0

		if 0 != expected {
			t.Errorf("Expected %d; but got %d", expected, g.GetNrNodes())
		}
	})

}
