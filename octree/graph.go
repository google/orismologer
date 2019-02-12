/*
Copyright 2019 Google LLC

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    https://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package octree

import (
	"fmt"
	"strings"
)

// Graph provides a basic public interface for graph types. It does not support multi-edges.
type Graph interface {
	// Nodes returns all nodes in the graph.
	// The result should have a stable order.
	Nodes() []string

	// Neighbors returns a list of neighbors (successors) to the given source node.
	// The result should have a stable order.
	Neighbors(a string) []string

	// Weight returns the weight associated with the given edge.
	Weight(a, z string) int64
}

type edge struct{ a, z string }

// AdjList is a mutable, directed, weighted graph implemented using an adjacency list.
// It can be treated as an unweighted graph by using AddEdge() which provides default weights of 1.
// It can be used as an undirected graph by using AddEdge() in each direction.
type AdjList struct {
	nodes  []string
	edges  map[string][]string
	weight map[edge]int64
}

// NewAdjList creates a new adjacency list graph.
func NewAdjList() *AdjList {
	return &AdjList{
		nodes:  []string{},
		edges:  make(map[string][]string),
		weight: make(map[edge]int64),
	}
}

// AddNode adds a node with the given name if it does not already exist.
func (g *AdjList) AddNode(n string) {
	if _, ok := g.edges[n]; ok {
		return
	}
	g.nodes = append(g.nodes, n)
	g.edges[n] = []string{}
}

// AddEdge adds an edge with the given name if it does not already exist.
// The default weight of 1 is used. If a graph is built only using AddEdge() it can be treated as
// unweighted graph in some algorithms below.
func (g *AdjList) AddEdge(a, z string) {
	g.AddWeightedEdge(a, z, 1)
}

// AddWeightedEdge adds an edge with the given name and weight if it does not already exist.
func (g *AdjList) AddWeightedEdge(a, z string, weight int64) {
	g.AddNode(a)
	g.AddNode(z)
	g.weight[edge{a, z}] = weight
	for _, n := range g.edges[a] {
		if z == n {
			return
		}
	}
	g.edges[a] = append(g.edges[a], z)
}

// Nodes returns all nodes in the graph.
func (g *AdjList) Nodes() []string {
	return g.nodes
}

// Edges returns all edges in the graph.
func (g *AdjList) Edges() map[string][]string {
	return g.edges
}

// Neighbors returns a list of neighbors to the given source node.
func (g *AdjList) Neighbors(a string) []string {
	return g.edges[a]
}

// Weight returns the edge weight for the given edge.
func (g *AdjList) Weight(a, z string) int64 {
	return g.weight[edge{a, z}]
}

// ToDot renders this graph to a dot format string, which can be helpful for debugging.
func (g *AdjList) ToDot() string {
	lines := []string{"digraph {"}
	for _, n := range g.Nodes() {
		lines = append(lines, fmt.Sprintf("\t%q", n))
	}
	for _, n := range g.Nodes() {
		for _, n2 := range g.Neighbors(n) {
			lines = append(lines, fmt.Sprintf("\t%q -> %q [label=%d]", n, n2, g.Weight(n, n2)))
		}
	}
	lines = append(lines, "}")
	return strings.Join(lines, "\n")
}
