package graph

import "sort"

type Graph struct {
	adj map[string][]string
}

func NewGraph() *Graph {
	return &Graph{
		adj: make(map[string][]string),
	}
}

func (g *Graph) AddNode(node string) {
	if _, exists := g.adj[node]; !exists {
		g.adj[node] = nil
	}
}

func (g *Graph) AddEdge(from, to string) {
	g.AddNode(from)
	g.AddNode(to)

	g.adj[from] = append(g.adj[from], to)
}

func (g *Graph) TopoSortLenient() []string {
	visited, temp := map[string]bool{}, map[string]bool{}

	var result []string

	nodes := make([]string, 0, len(g.adj))
	for node := range g.adj {
		nodes = append(nodes, node)
	}
	sort.Strings(nodes)

	var visit func(string)
	visit = func(node string) {
		if visited[node] {
			return
		}

		if temp[node] {
			return
		}

		temp[node] = true

		nbrs := append([]string(nil), g.adj[node]...)
		sort.Strings(nbrs)

		for _, m := range nbrs {
			visit(m)
		}

		temp[node], visited[node] = false, true
		result = append(result, node)
	}

	for _, node := range nodes {
		visit(node)
	}

	return result
}
