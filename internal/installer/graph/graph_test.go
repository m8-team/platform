package graph

import (
	"errors"
	"testing"
)

func TestTopologicalSortIsDeterministicByWaveThenID(t *testing.T) {
	graph, err := New([]Node{
		{ID: "apps", Wave: 0, Dependencies: []string{"gateway", "data"}},
		{ID: "crds", Wave: -100},
		{ID: "data", Wave: -50, Dependencies: []string{"crds"}},
		{ID: "gateway", Wave: -20, Dependencies: []string{"crds"}},
		{ID: "namespaces", Wave: -90, Dependencies: []string{"crds"}},
	})
	if err != nil {
		t.Fatalf("new graph: %v", err)
	}

	ordered, err := graph.Topological()
	if err != nil {
		t.Fatalf("topological: %v", err)
	}

	got := make([]string, 0, len(ordered))
	for _, node := range ordered {
		got = append(got, node.ID)
	}
	want := []string{"crds", "namespaces", "data", "gateway", "apps"}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("index %d: got %q want %q; full order %v", i, got[i], want[i], got)
		}
	}
}

func TestTopologicalSortDetectsCycle(t *testing.T) {
	graph, err := New([]Node{
		{ID: "a", Dependencies: []string{"b"}},
		{ID: "b", Dependencies: []string{"a"}},
	})
	if err != nil {
		t.Fatalf("new graph: %v", err)
	}

	_, err = graph.Topological()
	if !errors.Is(err, ErrDependencyCycle) {
		t.Fatalf("expected cycle error, got %v", err)
	}
}
