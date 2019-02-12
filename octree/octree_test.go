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
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/orismologer/utils"
)

func TestTreeBuildsMultiSegmentSubpathsCorrectly(t *testing.T) {
	tree := makeTree(t)
	children, err := tree.children("root/grandmother/aunt")
	if err != nil {
		t.Fatalf("%v", err)
	}
	if len(children) != 1 {
		t.Fatalf("expect aunt to have exactly one child, but got %v", len(children))
	}
	expected := "root/grandmother/aunt/cousin"
	got := children[0] // Make sure we don't get something like "root/grandmother/cousin"
	if got != expected {
		t.Fatalf("expected aunt's child to be named %q, but got %q", expected, got)
	}
}

func TestExpandPath(t *testing.T) {
	for _, test := range []struct {
		name     string
		path     string
		expected []string
	}{
		{
			name:     "just root",
			path:     "/",
			expected: []string{"root"},
		},
		{
			name:     "absolute path",
			path:     "/first/second/third",
			expected: []string{"root", "first", "second", "third"},
		},
		{
			name:     "relative path",
			path:     "first/second",
			expected: []string{"first", "second"},
		},
		{
			name:     "relative path, one node",
			path:     "first",
			expected: []string{"first"},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			got, err := expandPath(test.path)
			if err != nil {
				t.Errorf("%v: expected `%v`, got error: %v", test.name, test.expected, err)
			}
			if !cmp.Equal(test.expected, got) {
				t.Errorf("%v: expected `%v`, got `%v`", test.name, test.expected, got)
			}
		})
	}
}

func TestJoinPath(t *testing.T) {
	for _, test := range []struct {
		path     []string
		expected string
	}{
		{
			path:     []string{"root"},
			expected: "root",
		},
		{
			path:     []string{"root", "first", "second", "third"},
			expected: "root/first/second/third",
		},
		{
			path:     []string{"first", "second"},
			expected: "first/second",
		},
		{
			path:     []string{"first"},
			expected: "first",
		},
	} {
		testName := fmt.Sprintf("[%v]", strings.Join(test.path, ","))
		t.Run(testName, func(t *testing.T) {
			got := joinPath(test.path)
			if !cmp.Equal(test.expected, got) {
				t.Errorf("%v: expected `%v`, got `%v`", test.path, test.expected, got)
			}
		})
	}
}

func TestChildren(t *testing.T) {
	tree := makeTree(t)
	for _, test := range []struct {
		name     string
		node     string
		expected []string
	}{
		{
			name:     "root's children",
			node:     "root",
			expected: []string{"root/paternal_grandfather", "root/grandmother"},
		},
		{
			name:     "no children",
			node:     "root/paternal_grandfather/father/child",
			expected: []string{},
		},
		{
			name:     "children of a nested node",
			node:     "root/paternal_grandfather/father",
			expected: []string{"root/paternal_grandfather/father/child", "root/paternal_grandfather/father/sibling"},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			got, err := tree.children(test.node)
			if err != nil {
				t.Errorf("%v: expected `%v`, got error: %v", test.name, test.expected, err)
			}
			if !cmp.Equal(test.expected, got) {
				t.Errorf("%v: expected `%v`, got `%v`", test.name, test.expected, got)
			}
		})
	}
}

func TestIsValid(t *testing.T) {
	tree := makeTree(t)
	for _, test := range []struct {
		name              string
		path              string
		expectPathIsValid bool
	}{
		{
			name:              "just root",
			path:              "/",
			expectPathIsValid: true,
		},
		{
			name:              "does not contain invalid path",
			path:              "/invalid",
			expectPathIsValid: false,
		},
		{
			name:              "contains valid path",
			path:              "/paternal_grandfather",
			expectPathIsValid: true,
		},
		{
			name:              "contains valid nested path",
			path:              "/paternal_grandfather/father/child",
			expectPathIsValid: true,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			valid := tree.IsValid(test.path)
			switch {
			case test.expectPathIsValid && !valid:
				t.Errorf("%v: Expected valid, got invalid", test.path)
			case !test.expectPathIsValid && valid:
				t.Errorf("%v: Expected invalid, got valid", test.path)
			}
		})
	}
}

func TestNormalizePath(t *testing.T) {
	for _, test := range []struct {
		path          string
		expected      string
		expectedError bool
	}{
		{
			path:     "/",
			expected: "root",
		},
		{
			path:     "/first",
			expected: "root/first",
		},
		{
			path:     "/first/second",
			expected: "root/first/second",
		},
		{
			path:     "root",
			expected: "root",
		},
		{
			path:     "root/first",
			expected: "root/first",
		},
		{
			path:     "root/first/second",
			expected: "root/first/second",
		},
		{
			path:     "first",
			expected: "first",
		},
		{
			path:     "first/second",
			expected: "first/second",
		},
		{
			path:     "",
			expected: "",
		},
		{
			path:     "first/",
			expected: "first",
		},
		{
			path:     "first/second/",
			expected: "first/second",
		},
		{
			path:     "/first/",
			expected: "root/first",
		},
		{
			path:     "/first/second/",
			expected: "root/first/second",
		},
		{
			path:          "//",
			expectedError: true,
		},
		{
			path:          "first//second",
			expectedError: true,
		},
	} {
		t.Run(test.path, func(t *testing.T) {
			got, err := normalizePath(test.path)
			switch {
			case !test.expectedError && err != nil:
				t.Errorf("%v: expected `%v`, got error: %v", test.path, test.expected, err)
			case test.expectedError && err == nil:
				t.Errorf("%v: expected error, got `%v`", test.path, got)
			case got != test.expected:
				t.Errorf("%v: expected `%v`, got `%v`", test.path, test.expected, got)
			}
		})
	}
}

func TestGetTransformationIdentifier(t *testing.T) {
	tree := makeTree(t)
	for _, test := range []struct {
		name          string
		path          string
		expected      string
		expectedError bool
	}{
		{
			path:     "/grandmother/aunt/cousin",
			expected: "cousin_t",
		},
		{
			path:     "root/grandmother/aunt/cousin",
			expected: "cousin_t",
		},
		{
			path:          "invalid",
			expectedError: true,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			got, err := tree.GetTransformationIdentifier(test.path)
			switch {
			case !test.expectedError && err != nil:
				t.Errorf("GetTransformationIdentifier(%q): expected %q, got error: %v", test.path, test.expected, err)
			case test.expectedError && err == nil:
				t.Errorf("GetTransformationIdentifier(%q): expected error, got %q", test.path, got)
			case got != test.expected:
				t.Errorf("GetTransformationIdentifier(%q): expected %q, got %q", test.path, test.expected, got)
			}
		})
	}
}

func makeTree(t *testing.T) OcTree {
	const mappingsFile = "../testdata/oc_tree_test_mappings.pb"
	mappings, err := utils.LoadMappings(mappingsFile)
	if err != nil {
		t.Fatalf("Error during test set up: %v", err)
	}
	tree, err := NewTree(mappings)
	if err != nil {
		t.Fatalf("Error during test set up: %v", err)
	}
	return tree
}
