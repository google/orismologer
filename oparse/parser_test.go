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

package oparse

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name             string
		expressionString string
		expectedError    bool
	}{
		{
			name:          "empty expression",
			expectedError: true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := Parse(test.expressionString)
			switch {
			case err == nil && test.expectedError:
				t.Errorf("%v: expected error, but got no error", test.name)
			case err != nil && !test.expectedError:
				t.Errorf("%v: got error: %v", test.name, err)
			}
		})
	}
}

func TestEval(t *testing.T) {
	tests := []struct {
		name             string
		expressionString string
		context          Context
		expected         interface{}
		expectedError    bool
	}{
		// Arithmetic
		{
			name:             "arithmetic",
			expressionString: "1+2*3+4/2",
			expected:         9.0,
		},
		{
			name:             "brackets",
			expressionString: "2*(3+1)",
			expected:         8.0,
		},
		{
			name:             "arithmetic starting with brackets",
			expressionString: "(10 + 1) * 1000",
			expected:         11000.0,
		},
		{
			name:             "division by zero",
			expressionString: "100 / 0",
			expectedError:    true,
		},
		{
			name:             "indirect division by zero",
			expressionString: "100 / (1-1)",
			expectedError:    true,
		},

		// Variables
		{
			name:             "solo variable",
			expressionString: "i",
			context:          Context{"i": 10},
			expected:         10.0,
		},
		{
			name:             "arithmetic with a variable",
			expressionString: "i*2+3",
			context:          Context{"i": 10},
			expected:         23.0,
		},
		{
			name:             "invalid variable",
			expressionString: "j*2+3",
			context:          Context{"i": 10},
			expectedError:    true,
		},
		{
			name:             "variables starting with brackets",
			expressionString: "(boot_time + last_change_relative) * 1000",
			context:          Context{"boot_time": 10, "last_change_relative": 5},
			expected:         15000.0,
		},

		// Strings
		{
			name:             "string variable",
			expressionString: "i",
			context:          Context{"i": "hello"},
			expected:         "hello",
		},
		{
			name:             "string with single quotes",
			expressionString: "'hello world'",
			expected:         "hello world",
		},
		{
			name:             "string with double quotes",
			expressionString: "\"hello world\"",
			expected:         "hello world",
		},
		{
			name:             "string without quotes",
			expressionString: "hello",
			expectedError:    true,
		},
		{
			name:             "string with missing closing quote",
			expressionString: "'hello",
			expectedError:    true,
		},
		{
			name:             "string with missing opening quote",
			expressionString: "hello'",
			expectedError:    true,
		},
		{
			name:             "extraneous quote",
			expressionString: "'hello''",
			expectedError:    true,
		},
		{
			name:             "empty string",
			expressionString: "''",
			expected:         "",
		},
		{
			name:             "string concatenation",
			expressionString: "'hello' + ' ' + 'hello'",
			expected:         "hello hello",
		},
		{
			name:             "string variable concatenation",
			expressionString: "i+' '+i",
			context:          Context{"i": "hello"},
			expected:         "hello hello",
		},
		{
			name:             "String concatenation with numbers",
			expressionString: "'The answer is ' + 41 + 1",
			expected:         "The answer is 411",
		},
		{
			name:             "String concatenation with brackets",
			expressionString: "'The answer is ' + (41 + 1)",
			expected:         "The answer is 42",
		},
		{
			name:             "Invalid string concatenation",
			expressionString: "'The answer is ' * 2",
			expectedError:    true,
		},

		// Functions
		{
			name:             "function call",
			expressionString: "myfunc()",
			expected:         1.0,
		},
		{
			name:             "function call with missing closing bracket",
			expressionString: "myfunc(",
			expectedError:    true,
		},
		{
			name:             "function call with missing opening bracket",
			expressionString: "myfunc)",
			expectedError:    true,
		},
		{
			name:             "function call with single, numeric parameter",
			expressionString: "myfunc(100)",
			expected:         1.0,
		},
		{
			name:             "function call with single, string parameter",
			expressionString: "myfunc('hello')",
			expected:         1.0,
		},
		{
			name:             "function call with multiple parameters",
			expressionString: "myfunc(1, 'hello')",
			expected:         1.0,
		},
		{
			name:             "function call with variable parameter",
			expressionString: "myfunc(i)",
			context:          Context{"i": 999},
			expected:         1.0,
		},
		{
			name:             "function call with nested expressions",
			expressionString: "myfunc('hello'+' there', 9/3, anotherfunc(2+4))",
			expected:         1.0,
		},
		{
			name:             "function call with arithmetic",
			expressionString: "myfunc(100) + 3 * i / 5",
			context:          Context{"i": 10},
			expected:         7.0,
		},
		{
			name:             "function call starting with brackets",
			expressionString: "(boot_time + to_int(last_change_relative)) * 1000",
			context:          Context{"boot_time": 10, "last_change_relative": 5},
			expected:         11000.0,
		},
		{
			name:             "function call with string concatenation",
			expressionString: "'The answer is ' + (41 + myfunc(100))",
			expected:         "The answer is 42",
		},
	}
	// Dummy function caller which returns 1 for any function name.
	caller := func(funcName string, args ...interface{}) (interface{}, error) {
		return 1, nil
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			expression, err := Parse(test.expressionString)
			var got interface{}
			if err == nil {
				got, err = Eval(expression, test.context, caller)
			}
			switch {
			case !test.expectedError && err != nil:
				t.Errorf("%v: got `%v`, expected no error", test.name, err)
			case test.expectedError && err == nil:
				t.Errorf("%v: got no error, expected error", test.name)
			case !cmp.Equal(test.expected, got) && err == nil:
				t.Errorf("%v: got `%v`, expected `%v`", test.name, got, test.expected)
			}
		})
	}
}

func TestIdentifiers(t *testing.T) {
	tests := []struct {
		name             string
		expressionString string
		expectedVars     []string
		expectedFuncs    []string
	}{
		{
			name:             "no identifiers",
			expressionString: "1 + 3 - 4",
		},
		{
			name:             "one variable",
			expressionString: "i",
			expectedVars:     []string{"i"},
		},
		{
			name:             "one func",
			expressionString: "func()",
			expectedFuncs:    []string{"func"},
		},
		{
			name:             "one func, one var",
			expressionString: "i + func()",
			expectedFuncs:    []string{"func"},
			expectedVars:     []string{"i"},
		},
		{
			name:             "var in func",
			expressionString: "func(i)",
			expectedFuncs:    []string{"func"},
			expectedVars:     []string{"i"},
		},
		{
			name:             "arithmetic containing a var in a func",
			expressionString: "func(i+1)",
			expectedFuncs:    []string{"func"},
			expectedVars:     []string{"i"},
		},
		{
			name:             "complex",
			expressionString: "i + j + func(s, t) * myfunc(q + another(1+3))",
			expectedFuncs:    []string{"func", "myfunc", "another"},
			expectedVars:     []string{"i", "j", "s", "t", "q"},
		},
		{
			name:             "start with a bracket",
			expressionString: "(boot_time + to_int(last_change_relative)) * 1000",
			expectedFuncs:    []string{"to_int"},
			expectedVars:     []string{"boot_time", "last_change_relative"},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			expression, err := Parse(test.expressionString)
			gotVars, gotFuncs := expression.Identifiers()
			switch {
			case err != nil:
				t.Errorf("Identifiers(%q) got error: %v", test.expressionString, err)
			case !cmp.Equal(gotVars, test.expectedVars):
				t.Errorf("Identifiers(%q) got vars: %v; expected: %v", test.expressionString, gotVars, test.expectedVars)
			case !cmp.Equal(gotFuncs, test.expectedFuncs):
				t.Errorf("Identifiers(%q) got funcs: %v; expected: %v", test.expressionString, gotFuncs, test.expectedFuncs)
			}
		})
	}
}
