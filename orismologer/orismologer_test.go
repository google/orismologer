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

package orismologer

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/golang/glog"
	"github.com/google/go-cmp/cmp"
	"github.com/google/orismologer/utils"

	pb "github.com/google/orismologer/proto_out/proto"
)

func TestCanResolve(t *testing.T) {
	o, err := makeTestOrismologer()
	if err != nil {
		t.Fatalf("%v", err)
	}
	for _, test := range []struct {
		name     string
		nocPath  *pb.NocPath
		target   string
		expected bool
	}{
		{
			name: "eval-able",
			nocPath: &pb.NocPath{
				Oids: []string{"1.3.6.1.4.1.9.9.48.1.1.1.5.1"},
			},
			target:   "cisco",
			expected: true,
		},
		{
			name: "un-eval-able",
			nocPath: &pb.NocPath{
				Oids: []string{"1.3.6.1.4.1.9.9.48.1.1.1.5.1"},
			},
			target:   "aruba",
			expected: false,
		},
		{
			name: "invalid target",
			nocPath: &pb.NocPath{
				Oids: []string{"1.3.6.1.4.1.9.9.48.1.1.1.5.1"},
			},
			target:   "invalid",
			expected: false,
		},
		{
			name: "multiple OIDs",
			nocPath: &pb.NocPath{
				Oids: []string{
					"1.3.6.1.4.1.9.9.48.1.1.1.5.1",
					"1.3.6.1.4.1.9.9.48.1.1.1.5.2",
					"1.3.6.1.4.1.14823.2.2.1.2.1.6",
				},
			},
			target:   "aruba",
			expected: true,
		},
		{
			name: "standard MIB for Cisco target",
			nocPath: &pb.NocPath{
				Oids: []string{"1.3.6.1.2.1.25.3.3.1.2"},
			},
			target:   "cisco",
			expected: true,
		},
		{
			name: "standard MIB for Aruba target",
			nocPath: &pb.NocPath{
				Oids: []string{"1.3.6.1.2.1.25.3.3.1.2"},
			},
			target:   "aruba",
			expected: true,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			if got, want := o.canResolve(test.nocPath, test.target), test.expected; got != want {
				t.Errorf("canResolve() = %v, expected %v", got, want)
			}
		})
	}
}

func TestGetNocPaths(t *testing.T) {
	o, err := makeTestOrismologer()
	if err != nil {
		t.Fatalf("Could not set up test: %v", err)
	}
	for _, test := range []struct {
		name              string
		transformation    *pb.Transformation
		expectedPathNames []string
	}{
		{
			name: "",
			transformation: &pb.Transformation{
				Bind: "test",
				NocPaths: []*pb.NocPath{
					{Bind: "noc_path_1"},
					{},
					{Bind: "noc_path_3"},
				},
			},
			expectedPathNames: []string{"noc_path_1", "noc_path_3"},
		},
		{
			name: "",
			transformation: &pb.Transformation{
				Bind: "test",
				NocPaths: []*pb.NocPath{
					{Bind: "noc_path_1"},
					{Bind: "noc_path_2"},
					{Bind: "noc_path_3"},
				},
			},
			expectedPathNames: []string{"noc_path_1", "noc_path_2", "noc_path_3"},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			paths := o.getNocPaths(test.transformation)
			var pathNames []string
			for _, path := range paths {
				pathNames = append(pathNames, path.GetBind())
			}
			got := frequencyCounter(pathNames)
			expected := frequencyCounter(test.expectedPathNames)
			if !cmp.Equal(expected, got) {
				t.Fatalf("getNocPaths() = %v, expected %v", got, test.expectedPathNames)
			}
		})
	}
}

func TestEval(t *testing.T) {
	o, err := makeTestOrismologer()
	if err != nil {
		t.Fatalf("Could not set up test: %v", err)
	}
	for _, test := range []struct {
		transformationName string
		vendor             string
		expected           interface{}
		expectsError       bool
	}{
		{
			transformationName: "cpu_name",
			vendor:             "cisco",
			expectsError:       true,
		},
		{
			transformationName: "cpu_name",
			vendor:             "aruba",
			expected:           "Network Processor CPU10",
		},
		{
			transformationName: "boot_time",
			vendor:             "cisco",
			expected:           100.0,
		},
		{
			transformationName: "boot_time",
			vendor:             "aruba",
			expected:           100.0,
		},
		{
			transformationName: "last_change_absolute",
			vendor:             "cisco",
			expected:           150000.0,
		},
		{
			transformationName: "last_change_absolute",
			vendor:             "aruba",
			expected:           150000.0,
		},
	} {
		testName := test.transformationName + "_" + test.vendor
		t.Run(testName, func(t *testing.T) {
			transformation := o.transformations[test.transformationName]
			got, err := o.eval(transformation, "target", test.vendor)
			switch {
			case err != nil && !test.expectsError:
				t.Errorf("eval(), got error: %v", err)
			case err == nil && test.expectsError:
				t.Errorf("eval(), expected error, got: %v", got)
			case err == nil && !test.expectsError && !cmp.Equal(got, test.expected):
				t.Errorf("eval() = %v, expected: %v", got, test.expected)
			}
		})
	}
}

func makeTestOrismologer() (*Orismologer, error) {
	const transformationsFile = "../testdata/orismologer_test_transformations.pb"
	transformations, err := utils.LoadTransformations(transformationsFile)
	if err != nil {
		return nil, err
	}
	vendorInfo := &pb.VendorOids{
		VendorRoot: "1.3.6.1.4.1",
		Vendors: map[string]string{
			"cisco": "9",
			"aruba": "14823",
		},
	}
	o, err := newOrismologer(&pb.Mappings{}, transformations, vendorInfo)
	if err != nil {
		return &Orismologer{}, fmt.Errorf("could not create Orismologer: %v", err)
	}
	o.nocPathResolver = func(nocPath *pb.NocPath, target string) (interface{}, error) {
		samples := nocPath.GetSamples()
		if len(samples) != 1 {
			glog.Errorf("NocPath in test data should include exactly one sample")
			return nil, nil
		}
		return samples[0], nil
	}
	o.functions = dummyLibrary{}
	return o, nil
}

func frequencyCounter(strings []string) map[string]int {
	counters := map[string]int{}
	for _, s := range strings {
		counters[s]++
	}
	return counters
}

type dummyLibrary struct{}

func (l dummyLibrary) Call(funcName string, args ...interface{}) (interface{}, error) {
	switch funcName {
	case "to_int":
		i, _ := strconv.Atoi(args[0].(string))
		return i, nil
	case "to_string":
		return args[0].(string), nil
	case "time_since_epoch":
		return 20000100, nil
	default:
		return nil, fmt.Errorf("function %q undefined", funcName)
	}
}

func (l dummyLibrary) Contains(funcName string) (contains bool) {
	defer func() {
		if r := recover(); r != nil {
			contains = true
		}
	}()
	_, err := l.Call(funcName)
	contains = err == nil
	return contains
}
