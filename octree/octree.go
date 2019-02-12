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

// Package octree implements a tree data structure for storing OpenConfigNode protos.
package octree

import (
	"fmt"
	pb "github.com/google/orismologer/proto_out/proto"
	"strings"
)

const (
	pathSep = "/"

	// RootName is the name of the root node of an OcTree.
	RootName = "root"
)

/*
OcTree represents an OpenConfig tree. Nodes are represented as strings.
The underlying representation is a graph (adjacency list), with a map of node payloads.
*/
type OcTree struct {
	graph    *AdjList
	payloads map[string]*pb.OpenConfigNode
}

// NewTree creates and populates an OcTree from a Mappings proto.
func NewTree(mappings *pb.Mappings) (OcTree, error) {
	t := OcTree{
		graph:    NewAdjList(),
		payloads: map[string]*pb.OpenConfigNode{},
	}
	// Create a root OCNode so proto tree can be handled consistently.
	t.graph.AddNode(RootName)
	for _, node := range mappings.GetNodes() {
		if err := t.build(RootName, node); err != nil {
			return t, err
		}
	}
	return t, nil
}

/*
build recursively creates an OcTree given an OpenConfigNode proto and a relative or absolute path to
its parent. Ancestor nodes in the path will be created as needed.
*/
func (t *OcTree) build(parent string, current *pb.OpenConfigNode) error {
	subpath, err := expandPath(current.GetSubpath().GetPath())
	if err != nil {
		return err
	}

	// Root has no parent, so if given path starts at root (absolute) advance position by one node.
	if len(subpath) > 0 && subpath[0] == RootName {
		parent = RootName
		subpath = subpath[1:]
	}

	// Create each node in the given path, using its absolute path as the node name.
	fullPath := parent
	for _, node := range subpath {
		fullPath = fullPath + pathSep + node
		t.addChild(parent, fullPath)
		parent = fullPath
	}

	// Set the leaf node's payload (only these hold interesting data; others are just structure).
	if err := t.setPayload(fullPath, current); err != nil {
		return err
	}

	// Continue to build the tree, recursively, treating the current node as the parent.
	children := current.GetChildren()
	for _, child := range children {
		if err := t.build(fullPath, child); err != nil {
			return err
		}
	}
	return nil
}

func (t *OcTree) addChild(parent string, child string) error {
	if !t.IsValid(parent) {
		return fmt.Errorf("could not add child %q as parent %q does not exist", child, parent)
	}
	t.graph.AddEdge(parent, child)
	return nil
}

// children returns the child nodes of a node in an OcTree.
func (t *OcTree) children(parent string) ([]string, error) {
	if !t.IsValid(parent) {
		return nil, fmt.Errorf("could not get children of node with invalid path %q", parent)
	}
	return t.graph.Neighbors(parent), nil
}

func (t *OcTree) getPayload(node string) (*pb.OpenConfigNode, error) {
	if !t.IsValid(node) {
		return nil, fmt.Errorf("could not get payload as no such node in tree: %q", node)
	}
	if val, ok := t.payloads[node]; ok {
		return val, nil
	}
	return nil, fmt.Errorf("payload missing for node %q", node)
}

func (t *OcTree) setPayload(node string, payload *pb.OpenConfigNode) error {
	if !t.IsValid(node) {
		return fmt.Errorf("could not set payload as no such node in tree: %q", node)
	}
	t.payloads[node] = payload
	return nil
}

/*
IsValid returns true if a given OpenConfig path is defined in the OcTree.
Paths are given as "root/parent/child" or, equivalently, as "/parent/child".
*/
func (t *OcTree) IsValid(path string) bool {
	path, err := normalizePath(path)
	if err != nil {
		return false
	}
	_, ok := t.graph.Edges()[path]
	return ok
}

// GetTransformationIdentifier returns the identifier of the transformation for a given OC path.
func (t *OcTree) GetTransformationIdentifier(path string) (string, error) {
	node, err := normalizePath(path)
	if err != nil {
		return "", err
	}
	payload, err := t.getPayload(node)
	if err != nil {
		return "", err
	}
	return payload.GetBind(), nil
}

// Print pretty prints a subtree rooted at the given node.
func (t *OcTree) Print(root string) error {
	if !t.IsValid(root) {
		return fmt.Errorf("cannot print tree from nonexistant node %q", root)
	}
	return t._printTree(root, root, "", false)
}

func (t *OcTree) _printTree(originalRoot string, current string, prefix string, last bool) error {
	originalRoot, err := normalizePath(originalRoot)
	if err != nil {
		return fmt.Errorf("could not print tree: %v", err)
	}
	current, err = normalizePath(current)
	if err != nil {
		return fmt.Errorf("could not print tree: %v", err)
	}
	path, err := expandPath(current)
	if err != nil {
		return fmt.Errorf("could not print tree: %v", err)
	}
	nodeName := path[len(path)-1]

	fmt.Print(prefix)
	switch {
	case last:
		fmt.Print("└── ")
		prefix = fmt.Sprintf("%v    ", prefix)
	case current != originalRoot:
		fmt.Print("├── ")
		prefix = fmt.Sprintf("%v|   ", prefix)
	}
	fmt.Println(nodeName)

	children, err := t.children(current)
	if err != nil {
		return fmt.Errorf("could not print tree: %v", err)
	}
	for i, child := range children {
		t._printTree(originalRoot, child, prefix, i == len(children)-1)
	}
	return nil
}

/*
expandPath takes a path string and returns it, normalized, as an array of path segments.
eg: "/path/to/something" -> [root path to something]
*/
func expandPath(path string) ([]string, error) {
	path, err := normalizePath(path)
	if err != nil {
		return nil, fmt.Errorf("could not expand path: %v", err)
	}
	return strings.Split(path, pathSep), nil
}

func joinPath(path []string) string {
	return strings.Join(path, pathSep)
}

/*
Normalize path accepts path strings and returns the canonical representation used internally in this
package. Eg:

/first/second      ->  root/first/second (expand `/` to `root/`)
root/first/second  ->  root/first/second (no change)
first/second       ->  first/second      (relative path, so no `root` is added)

It also removes trailing slashes, eg: `first/second/` becomes `first/second`.
*/
func normalizePath(path string) (string, error) {
	if path == pathSep {
		return RootName, nil
	}
	if strings.Contains(path, pathSep+pathSep) {
		return "", fmt.Errorf("invalid path %q", path)
	}
	if strings.HasSuffix(path, pathSep) {
		path = strings.TrimSuffix(path, pathSep)
	}
	if strings.HasPrefix(path, pathSep) {
		return RootName + path, nil
	}
	return path, nil
}
