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

import (
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/golang/protobuf/proto"

	pb "github.com/google/orismologer/proto_out/proto"
)

// LoadMappings deserializes a text proto file at a given path as a Mappings proto message.
func LoadMappings(mappingsFile string) (*pb.Mappings, error) {
	bytes, err := ioutil.ReadFile(mappingsFile)
	if err != nil {
		return nil, fmt.Errorf("could not open mappings file: %v", err)
	}
	mappings := &pb.Mappings{}
	if err := proto.UnmarshalText(string(bytes), mappings); err != nil {
		return nil, fmt.Errorf("could not deserialize mappings: %v", err)
	}
	return mappings, nil
}

// LoadTransformations deserializes a text proto file at a given path as a Transformations proto
// message.
func LoadTransformations(transformationsFile string) (*pb.Transformations, error) {
	bytes, err := ioutil.ReadFile(transformationsFile)
	if err != nil {
		return nil, fmt.Errorf("could not open transformations file: %v", err)
	}
	transformations := &pb.Transformations{}
	if err := proto.UnmarshalText(string(bytes), transformations); err != nil {
		return nil, fmt.Errorf("could not deserialize transformations: %v", err)
	}
	return transformations, nil
}

// LoadVendorOids deserializes a text proto file at a given path as a VendorOids proto message.
func LoadVendorOids(vendorOidsFile string) (*pb.VendorOids, error) {
	bytes, err := ioutil.ReadFile(vendorOidsFile)
	if err != nil {
		return nil, fmt.Errorf("could not open vendor OIDs file: %v", err)
	}
	vendorOids := &pb.VendorOids{}
	if err := proto.UnmarshalText(string(bytes), vendorOids); err != nil {
		return nil, fmt.Errorf("could not deserialize vendor OIDs: %v", err)
	}
	return vendorOids, nil
}

// SliceToString returns a comma-separated string representing the contents of a slice.
func SliceToString(slice []interface{}) string {
	valueStrings := make([]string, len(slice))
	for i, value := range slice {
		valueStrings[i] = fmt.Sprint(value)
	}
	return strings.Join(valueStrings, ", ")
}
