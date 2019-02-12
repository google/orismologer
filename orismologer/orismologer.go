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

/*
Package orismologer translates non-OpenConfig telemetry sources (eg: SNMP OIDs) to OpenConfig paths.
*/
package orismologer

import (
	"fmt"
	"strings"

	"github.com/golang/glog"
	"github.com/google/orismologer/functions"
	"github.com/google/orismologer/octree"
	"github.com/google/orismologer/oparse"
	"github.com/google/orismologer/utils"

	pb "github.com/google/orismologer/proto_out/proto"
)

type transformationMap map[string]*pb.Transformation
type nocPathResolver func(*pb.NocPath, string) (interface{}, error)
type functionLibrary interface {
	Contains(funcName string) bool
	Call(funcName string, args ...interface{}) (interface{}, error)
}

// Orismologer translates non-OpenConfig telemetry sources (eg: SNMP OIDs) to OpenConfig paths.
type Orismologer struct {
	mappings        octree.OcTree
	transformations transformationMap
	vendorInfo      *pb.VendorOids
	nocPathResolver nocPathResolver
	functions       functionLibrary
}

/*
NewOrismologer builds an Orismologer instance from the text protos in the given files.
mappingsFile should contain a Mappings proto.
transformationFile should contain a Transformations proto.
vendorOidsFile should contain a VendorOids proto.
*/
func NewOrismologer(mappingsFile, transformationsFile, vendorOidsFile string) (*Orismologer, error) {
	mappings, err := utils.LoadMappings(mappingsFile)
	if err != nil {
		return nil, err
	}
	transformations, err := utils.LoadTransformations(transformationsFile)
	if err != nil {
		return nil, err
	}
	vendorOids, err := utils.LoadVendorOids(vendorOidsFile)
	if err != nil {
		return nil, err
	}
	return newOrismologer(mappings, transformations, vendorOids)
}

func newOrismologer(mappings *pb.Mappings, transformations *pb.Transformations, vendorInfo *pb.VendorOids) (*Orismologer, error) {
	t, err := octree.NewTree(mappings)
	if err != nil {
		return nil, err
	}
	transformationMap, err := makeTransformationMap(transformations)
	if err != nil {
		return nil, err
	}
	return &Orismologer{
		mappings:        t,
		transformations: transformationMap,
		vendorInfo:      vendorInfo,
		nocPathResolver: resolve,
		functions:       functions.NewLibrary(),
	}, nil
}

func makeTransformationMap(transformations *pb.Transformations) (transformationMap, error) {
	transformationMap := transformationMap{}
	for _, transformation := range transformations.GetTransformations() {
		name := transformation.GetBind()
		if _, ok := transformationMap[name]; ok {
			return nil, fmt.Errorf("more than one transformation bound to identifier %q", name)
		}
		transformationMap[name] = transformation
	}
	return transformationMap, nil
}

// PrintOcPaths pretty prints the tree of OpenConfig paths defined for this Orismologer instance.
func (o *Orismologer) PrintOcPaths(root string) error {
	return o.mappings.Print(root)
}

/*
Eval retrieves the current value of a given OpenConfig path for a target which does not natively
support OpenConfig.
The vendor name is used to identify dependencies for the target (eg: which OIDs it supports).
*/
// TODO: Support a dry run, to validate mappings and transformations protos.
func (o *Orismologer) Eval(openConfigPath, target, vendor string) (interface{}, error) {
	transformationName, err := o.mappings.GetTransformationIdentifier(openConfigPath)
	if err != nil {
		return nil, fmt.Errorf("failed to identify a transformation for path %q: %v", openConfigPath, err)
	}
	transformation, ok := o.transformations[transformationName]
	if !ok {
		return nil, fmt.Errorf("could not locate transformation %q for path %q", transformationName, openConfigPath)
	}
	glog.Infof("found transformation %q for path %q", transformationName, openConfigPath)
	return o.eval(transformation, target, vendor)
}

/*
eval parses and evaluates a Transformation proto's Expressions field, resolving any variables used
in expressions to their associated Transformations and recursively evaluating those until a final
value is obtained by resolving a NocPath. If a transformation defines multiple expressions then the
output of the first one that successfully evaluates is returned.

NocPaths are resolved using the function given to the Orismologer instance at instantiation.
*/
// TODO: Eval paths with keys, eg: thing/name[name=value]
// TODO: Safeguard against really long paths, and circular references.
func (o *Orismologer) eval(transformation *pb.Transformation, target string, vendor string) (interface{}, error) {
	transformationName := transformation.GetBind()
	glog.Infof("evaluating transformation %q for target %q of vendor %q", transformationName, target, vendor)
	nocPaths := o.getNocPaths(transformation)
	// Try to eval each expression defined for this transformation, taking the first that works.
	for _, expressionString := range transformation.GetExpressions() {
		glog.Infof("evaluating expression `%v`", expressionString)
		expression, variables, _, err := o.parseAndValidateExpression(expressionString)
		if err != nil {
			glog.Errorf("%v", err)
			continue
		}
		values, err := o.evalVariables(variables, nocPaths, target, vendor)
		if err != nil {
			if unresolvableNocPathError, ok := err.(unresolvableNocPathError); ok {
				glog.Info(unresolvableNocPathError.msg) // This is not an error we need to surface to the user.
			} else {
				glog.Errorf("%v", err)
			}
			glog.Infof("could not evaluate all variables for expression `%v`, continuing to next expression", expressionString)
			continue
		}

		// Evaluate the expression, passing in the values of the variables it uses.
		transformationResult, err := oparse.Eval(expression, values, o.functions.Call)
		if err != nil {
			return nil, err
		}
		return transformationResult, nil
	}
	return nil, fmt.Errorf("none of the expressions of transformation %q could be evaluated (see logs for details)", transformationName)
}

// getNocPaths returns a map of all the NocPaths defined in the given transformation.
func (o *Orismologer) getNocPaths(transformation *pb.Transformation) map[string]*pb.NocPath {
	transformationName := transformation.GetBind()
	paths := map[string]*pb.NocPath{}
	for _, nocPath := range transformation.GetNocPaths() {
		pathName := nocPath.GetBind()
		if len(pathName) == 0 {
			glog.Errorf("Transformation %q contains a NocPath without an identifier", transformationName)
		} else {
			glog.Infof("storing NocPath %q of transformation %q", pathName, transformationName)
			paths[pathName] = nocPath
		}
	}
	return paths
}

/*
Returns the expression parsed from the given string and any variables and function names used in it.
*/
func (o *Orismologer) parseAndValidateExpression(expressionString string) (*oparse.Expression, []string, []string, error) {
	expression, err := oparse.Parse(expressionString)
	if err != nil {
		glog.Errorf("could not parse expression `%v`", expressionString)
		return nil, nil, nil, err
	}
	variables, functionNames := expression.Identifiers()
	for _, functionName := range functionNames {
		if !o.functions.Contains(functionName) {
			return nil, nil, nil, fmt.Errorf("function %q is not defined", functionName)
		}
	}
	return expression, variables, functionNames, nil
}

/*
Evaluates each of the given variables, returning an error if one or more cannot be evaluated.
*/
func (o *Orismologer) evalVariables(variables []string, nocPaths map[string]*pb.NocPath, target string, vendor string) (map[string]interface{}, error) {
	values := oparse.Context{}
	for _, variable := range variables {
		glog.Infof("evaluating variable %q", variable)
		var value interface{}
		var err error
		nocPath := nocPaths[variable]
		transformation := o.transformations[variable]
		switch {
		case nocPath != nil:
			value, err = o.handleNocPath(nocPath, target, vendor)
			if err != nil {
				return nil, err
			}
		case transformation != nil:
			value, err = o.eval(transformation, target, vendor)
			if err != nil {
				return nil, fmt.Errorf("could not evaluate sub-transformation %q: %v", variable, err)
			}
		default:
			return nil, fmt.Errorf("NocPath or sub-transformation %q is undefined", variable)
		}
		glog.Infof("evaluated variable %q = %v", variable, value)
		values[variable] = value
	}
	return values, nil
}

// Gets a value for the given NocPath for the given target.
func (o *Orismologer) handleNocPath(nocPath *pb.NocPath, target string, vendor string) (interface{}, error) {
	pathName := nocPath.GetBind()
	if !o.canResolve(nocPath, vendor) {
		return nil, unresolvableNocPathError{
			fmt.Sprintf("ignoring NocPath %q as it cannot be resolved for vendor %q", pathName, vendor),
		}
	}
	value, err := o.nocPathResolver(nocPath, target)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve NocPath %q for target %q (this NocPath should normally be resolvable for this target): %v", pathName, target, err)
	}
	return value, nil
}

type unresolvableNocPathError struct {
	msg string
}

func (f unresolvableNocPathError) Error() string {
	return f.msg
}

// canResolve returns true if the given target supports the given NocPath.
func (o *Orismologer) canResolve(nocPath *pb.NocPath, vendor string) bool {
	// NB: Currently assumes NocPaths are OIDs only.
	vendorRoot := o.vendorInfo.GetVendorRoot()
	for _, oid := range nocPath.GetOids() {
		if !strings.HasPrefix(oid, vendorRoot) {
			return true
		}
		vendorOid, ok := o.vendorInfo.GetVendors()[vendor]
		if !ok {
			return false
		}
		if strings.HasPrefix(oid, vendorRoot+"."+vendorOid) {
			return true
		}
	}
	return false
}

/*
resolve retrieves the value for a given NocPath from a given target.
This may involve sending an SNMP request, running a CLI command and parsing the output, etc.
*/
func resolve(nocPath *pb.NocPath, target string) (interface{}, error) {
	// TODO: Implement.
	glog.Infof("Requesting NocPath %q from target %q", nocPath.GetBind(), target)
	samples := nocPath.GetSamples()
	if len(samples) > 0 {
		return samples[0], nil
	}
	return "dummy", nil
}
