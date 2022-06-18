package graph

import (
	"testing"
)

func TestAdd(t *testing.T) {

	// TODO: add test -> examples in repo below:
	// https://github.com/networkx/networkx/blob/main/networkx/classes/tests/test_graph.py

	t.Run("Empty Graph", func(t *testing.T) {
		g := new(CGraph)
		expected := 0

		if expected != g.GetNrNodes() {
			t.Errorf("Expected %d; but got %d", expected, g.GetNrNodes())
		}
	})

}
