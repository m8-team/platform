package graph

import (
	"errors"
	"fmt"
	"sort"
)

var (
	ErrDuplicateNode     = errors.New("duplicate graph node")
	ErrMissingDependency = errors.New("missing graph dependency")
	ErrDependencyCycle   = errors.New("dependency cycle")
)

type Node struct {
	ID           string
	Wave         int
	Dependencies []string
}

type Graph struct {
	nodes map[string]Node
}

func New(nodes []Node) (Graph, error) {
	graph := Graph{nodes: make(map[string]Node, len(nodes))}
	for _, node := range nodes {
		if node.ID == "" {
			return Graph{}, fmt.Errorf("node id is required")
		}
		if _, ok := graph.nodes[node.ID]; ok {
			return Graph{}, fmt.Errorf("%w: %s", ErrDuplicateNode, node.ID)
		}
		graph.nodes[node.ID] = normalizeNode(node)
	}

	for _, node := range graph.nodes {
		for _, dependency := range node.Dependencies {
			if _, ok := graph.nodes[dependency]; !ok {
				return Graph{}, fmt.Errorf("%w: %s depends on %s", ErrMissingDependency, node.ID, dependency)
			}
		}
	}

	return graph, nil
}

func normalizeNode(node Node) Node {
	out := node
	out.Dependencies = append([]string(nil), node.Dependencies...)
	sort.Strings(out.Dependencies)
	return out
}

func (g Graph) Topological() ([]Node, error) {
	indegree := make(map[string]int, len(g.nodes))
	dependents := make(map[string][]string, len(g.nodes))
	for id := range g.nodes {
		indegree[id] = 0
	}
	for _, node := range g.nodes {
		for _, dependency := range node.Dependencies {
			indegree[node.ID]++
			dependents[dependency] = append(dependents[dependency], node.ID)
		}
	}
	for dependency := range dependents {
		sortNodeIDs(g.nodes, dependents[dependency])
	}

	ready := make([]string, 0, len(g.nodes))
	for id, degree := range indegree {
		if degree == 0 {
			ready = append(ready, id)
		}
	}
	sortNodeIDs(g.nodes, ready)

	ordered := make([]Node, 0, len(g.nodes))
	for len(ready) > 0 {
		id := ready[0]
		ready = ready[1:]
		node := g.nodes[id]
		ordered = append(ordered, node)

		for _, dependent := range dependents[id] {
			indegree[dependent]--
			if indegree[dependent] == 0 {
				ready = append(ready, dependent)
			}
		}
		sortNodeIDs(g.nodes, ready)
	}

	if len(ordered) != len(g.nodes) {
		return nil, ErrDependencyCycle
	}

	return ordered, nil
}

func sortNodeIDs(nodes map[string]Node, ids []string) {
	sort.SliceStable(ids, func(i, j int) bool {
		left := nodes[ids[i]]
		right := nodes[ids[j]]
		if left.Wave != right.Wave {
			return left.Wave < right.Wave
		}
		return left.ID < right.ID
	})
}
