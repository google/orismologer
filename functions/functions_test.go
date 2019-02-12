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

package functions

import (
	"fmt"
	"strings"
	"testing"
)

func TestLibraryCall(t *testing.T) {
	l := makeDummyLibrary()
	tests := []struct {
		name         string
		funcName     string
		args         []interface{}
		expected     interface{}
		expectsError bool
	}{
		{
			name:     "valid call",
			funcName: "dummy",
			args:     []interface{}{"test"},
			expected: "test",
		},
		{
			name:         "undefined function",
			funcName:     "undefined",
			expectsError: true,
		},
		{
			name:         "too few args",
			funcName:     "dummy",
			expectsError: true,
		},
		{
			name:         "too many args",
			funcName:     "dummy",
			expectsError: true,
		},
		{
			name:     "one output",
			funcName: "oneOutput",
			expected: "1",
		},
		{
			name:         "too many outputs",
			funcName:     "threeOutputs",
			expectsError: true,
		},
		{
			name:         "not enough outputs",
			funcName:     "noOutputs",
			expectsError: true,
		},
		{
			name:         "second output not error or nil",
			funcName:     "secondOutputNotError",
			expectsError: true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := l.Call(test.funcName, test.args...)
			argStrings := []string{fmt.Sprintf("%q", test.funcName)}
			for _, arg := range test.args {
				argStrings = append(argStrings, fmt.Sprint(arg))
			}
			argString := strings.Join(argStrings, ", ")
			switch {
			case err != nil && !test.expectsError:
				t.Errorf("call(%v) expected %v, got error: %v", argString, test.expected, err)
			case err == nil && test.expectsError:
				t.Errorf("call(%v) got: %v, expected error", argString, got)
			case err == nil && got != test.expected:
				t.Errorf("call(%v) = %v, expected: %v", argString, got, test.expected)
			}
		})
	}
}

func TestLibraryToInt(t *testing.T) {
	tests := []struct {
		name         string
		input        interface{}
		expected     int
		expectsError bool
	}{
		{
			name:     "valid call",
			input:    10,
			expected: 10,
		},
		{
			name:     "string",
			input:    "10",
			expected: 10,
		},
		{
			name:         "float",
			input:        10.0,
			expectsError: true,
		},
		{
			name:         "float string",
			input:        "10.0",
			expectsError: true,
		},
		{
			name:         "overflow",
			input:        "999999999999999999999999999",
			expectsError: true,
		},
		{
			name:     "negative",
			input:    -1,
			expected: -1,
		},
		{
			name:         "slice",
			input:        []int{10},
			expectsError: true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := toInt(test.input)
			switch {
			case err != nil && !test.expectsError:
				t.Errorf("toInt(%v) expected %v, got error: %v", test.input, test.expected, err)
			case err == nil && test.expectsError:
				t.Errorf("toInt(%v) got: %v, expected error", test.input, got)
			case err == nil && got != test.expected:
				t.Errorf("toInt(%v) = %v, expected: %v", test.input, got, test.expected)
			}
		})
	}
}

func TestLibraryToStr(t *testing.T) {
	tests := []struct {
		name         string
		input        interface{}
		expected     string
		expectsError bool
	}{
		{
			name:     "plain string",
			input:    "thing",
			expected: "thing",
		},
		{
			name:         "integer",
			input:        10,
			expectsError: true,
		},
		{
			name:         "float",
			input:        10.0,
			expectsError: true,
		},
		{
			name:     "integer string",
			input:    "10",
			expected: "10",
		},
		{
			name:     "float string",
			input:    "10.0",
			expected: "10.0",
		},
		{
			name:         "slice",
			input:        []string{"string"},
			expectsError: true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := toStr(test.input)
			switch {
			case err != nil && !test.expectsError:
				t.Errorf("toInt(%v) expected %v, got error: %v", test.input, test.expected, err)
			case err == nil && test.expectsError:
				t.Errorf("toInt(%v) got: %v, expected error", test.input, got)
			case err == nil && got != test.expected:
				t.Errorf("toInt(%v) = %v, expected: %v", test.input, got, test.expected)
			}
		})
	}
}

func TestLibraryToFloat(t *testing.T) {
	tests := []struct {
		name         string
		input        interface{}
		expected     float64
		expectsError bool
	}{
		{
			name:     "float",
			input:    10.0,
			expected: 10.0,
		},
		{
			name:     "float string",
			input:    "10.0",
			expected: 10.0,
		},
		{
			name:         "integer",
			input:        10,
			expectsError: true,
		},
		{
			name:     "integer string",
			input:    "10",
			expected: 10.0,
		},
		{
			name:     "negative",
			input:    -1.0,
			expected: -1.0,
		},
		{
			name:         "slice",
			input:        []int{-10.0},
			expectsError: true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := toFloat(test.input)
			switch {
			case err != nil && !test.expectsError:
				t.Errorf("toFloat(%v) expected %v, got error: %v", test.input, test.expected, err)
			case err == nil && test.expectsError:
				t.Errorf("toFloat(%v) got: %v, expected error", test.input, got)
			case err == nil && got != test.expected:
				t.Errorf("toFloat(%v) = %v, expected: %v", test.input, got, test.expected)
			}
		})
	}
}

func TestLibraryTimeSinceEpoch(t *testing.T) {
	tests := []struct {
		name         string
		timeStamp    interface{}
		format       string
		units        string
		expected     int
		expectsError bool
	}{
		{
			name:      "ntp with spaces to s",
			timeStamp: "dfc4 0b68 8147 af78",
			format:    "ntp",
			units:     "s",
			expected:  1545178344,
		},
		{
			name:      "ntp without spaces to s",
			timeStamp: "dfc40b688147af78",
			format:    "ntp",
			units:     "s",
			expected:  1545178344,
		},
		{
			name:      "ntp to ms",
			timeStamp: "dfc40b688147af78",
			format:    "ntp",
			units:     "ms",
			expected:  1545178344505,
		},
		{
			name:      "ntp to ns",
			timeStamp: "dfc40b688147af78",
			format:    "ntp",
			units:     "ns",
			expected:  1545178344505000082,
		},
		{
			name:      "iso8601 to s",
			timeStamp: "2018-12-19 00:12:24",
			format:    "2006-01-02 15:04:05",
			units:     "s",
			expected:  1545178344,
		},
		{
			name:      "RFC3339 to s",
			timeStamp: "2018-12-19T11:12:24+11:00",
			format:    "rfc3339",
			units:     "s",
			expected:  1545178344,
		},
		{
			name:         "missing format",
			timeStamp:    "2018-12-18 15:15:59",
			units:        "s",
			expectsError: true,
		},
		{
			name:         "missing units",
			timeStamp:    "2018-12-18 15:15:59",
			format:       "iso8601",
			expectsError: true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := timeSinceEpoch(test.timeStamp, test.format, test.units)
			switch {
			case err != nil && !test.expectsError:
				t.Errorf("timeSinceEpoch(%q, %q, %q) expected %v, got error: %v", test.timeStamp, test.format, test.units, test.expected, err)
			case err == nil && test.expectsError:
				t.Errorf("timeSinceEpoch(%q, %q, %q) got: %v, expected error", test.timeStamp, test.format, test.units, got)
			case err == nil && got != test.expected:
				t.Errorf("timeSinceEpoch(%q, %q, %q) = %v, expected: %v", test.timeStamp, test.format, test.units, got, test.expected)
			}
		})
	}
}

func makeDummyLibrary() Library {
	registry := map[string]interface{}{
		"dummy":                dummy,
		"noOutputs":            noOutputs,
		"threeOutputs":         threeOutputs,
		"oneOutput":            oneOutput,
		"secondOutputNotError": secondOutputNotError,
	}
	return newLibrary(registry)
}

func dummy(arg string) string {
	return arg
}

func noOutputs() {
}

func oneOutput() string {
	return "1"
}

func threeOutputs() (string, string, string) {
	return "1", "2", "2"
}

func secondOutputNotError() (string, string) {
	return "1", "2"
}
