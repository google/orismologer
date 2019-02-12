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

// Package utils provides miscellaneous utilities for Orismologer.
package utils

import "testing"

func TestSliceToString(t *testing.T) {
	for _, test := range []struct {
		name     string
		slice    []interface{}
		expected string
	}{
		{
			name:     "simple",
			slice:    []interface{}{"one", "two"},
			expected: "one, two",
		},
		{
			name:     "empty",
			slice:    []interface{}{},
			expected: "",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			if got := SliceToString(test.slice); got != test.expected {
				t.Errorf("SliceToString() = %v, expected %v", got, test.expected)
			}
		})
	}
}
