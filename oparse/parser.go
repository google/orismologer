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
Package oparse parses simple expressions in orismologer protos.
Basic arithmetic, variables, function calls, string literals, nested expressions, and string
concatenation are supported.
Based on the version originally published at:
https://github.com/alecthomas/participle/blob/master/_examples/expr/main.go
*/
package oparse

import (
	"errors"
	"fmt"
	"log"
	"math"
	"strings"

	"github.com/alecthomas/participle"
	"github.com/golang/glog"
)

// Operator represents an arithmetic (or string interpolation) operator, eg: +.
type Operator int

const (
	// OpMul represents a multiplication symbol (*).
	OpMul Operator = iota

	// OpDiv represents a division symbol (/).
	OpDiv

	// OpAdd represents an addition symbol (+).
	OpAdd

	// OpSub represents a subtraction symbol (-).
	OpSub
)

var operatorMap = map[string]Operator{"+": OpAdd, "-": OpSub, "*": OpMul, "/": OpDiv}

// Capture implements Participle's Capture interface.
func (o *Operator) Capture(s []string) error {
	*o = operatorMap[s[0]]
	return nil
}

// Arg captures a function argument as an identifier optionally followed by a comma.
type Arg struct {
	Value     Expression `@@` // nolint: govet
	Separator *string    `[ "," ]`
}

// Function captures a function call as an identifier followed by a matched pair of brackets which
// contain 0 or more arguments.
type Function struct {
	Name  string `@Ident`
	Open  string `"("`
	Args  []*Arg `{ @@ }`
	Close string `")"`
}

// Value captures a value, which is either a literal of some kind (eg: a string or a number) or
// something that evaluates to one (eg: a function call, or a nested expression).
type Value struct {
	// NB: All numeric values will be represented as floats, to simplify parsing.
	Number        *float64    `@(Float|Int)`
	StrLiteral    *string     `| @(String|Char)`
	Function      *Function   `| @@`
	Variable      *string     `| @Ident`
	Subexpression *Expression `| "(" @@ ")"`
}

// Factor captures a base and an exponent.
type Factor struct {
	Base     *Value `@@`
	Exponent *Value `[ "^" @@ ]`
}

// OpFactor captures a multiplication or division operator followed by a factor.
type OpFactor struct {
	Operator Operator `@("*" | "/")`
	Factor   *Factor  `@@`
}

// Term captures a Factor followed by an OpFactor.
type Term struct {
	Left  *Factor     `@@`
	Right []*OpFactor `{ @@ }`
}

// OpTerm captures a plus or minus operator followed by a term.
type OpTerm struct {
	Operator Operator `@("+" | "-")`
	Term     *Term    `@@`
}

// Expression is the top level node in the grammar AST. It represents the complete expression to be
// parsed and evaluated.
type Expression struct {
	Left  *Term     `@@`
	Right []*OpTerm `{ @@ }`
}

// Functions for displaying parsed expressions. Useful for debugging.

func (o Operator) String() string {
	switch o {
	case OpMul:
		return "*"
	case OpDiv:
		return "/"
	case OpSub:
		return "-"
	case OpAdd:
		return "+"
	}
	glog.Error("Got unsupported operator while parsing expression")
	return "?"
}

func (f *Function) String() string {
	var args []string
	for _, arg := range f.Args {
		args = append(args, arg.Value.String())
	}
	return fmt.Sprintf("%v(%v)", f.Name, strings.Join(args, ", "))
}

func (v *Value) String() string {
	switch {
	case v.Number != nil:
		return fmt.Sprintf("%g", *v.Number)
	case v.StrLiteral != nil:
		return fmt.Sprintf("%q", *v.StrLiteral)
	case v.Variable != nil:
		return *v.Variable
	case v.Function != nil:
		return v.Function.String()
	case v.Subexpression != nil:
		return "(" + v.Subexpression.String() + ")"
	default:
		return ""
	}
}

func (f *Factor) String() string {
	out := f.Base.String()
	if f.Exponent != nil {
		out += " ^ " + f.Exponent.String()
	}
	return out
}

func (o *OpFactor) String() string {
	return fmt.Sprintf("%s %s", o.Operator, o.Factor)
}

func (t *Term) String() string {
	out := []string{t.Left.String()}
	for _, r := range t.Right {
		out = append(out, r.String())
	}
	return strings.Join(out, " ")
}

func (o *OpTerm) String() string {
	return fmt.Sprintf("%s %s", o.Operator, o.Term)
}

func (e *Expression) String() string {
	out := []string{e.Left.String()}
	for _, r := range e.Right {
		out = append(out, r.String())
	}
	return strings.Join(out, " ")
}

// Functions for actually evaluating parsed expressions.

func (o Operator) eval(l, r interface{}) (interface{}, error) {
	_, lIsInt := l.(int)
	_, rIsInt := r.(int)
	// Because of earlier handling we can assume that all numeric values are represented as floats.
	// ie: We should never get an int here.
	if lIsInt || rIsInt {
		log.Fatal("Evaluated parser output contained an int. That should not have happened.")
	}

	lFloat, lIsFloat := l.(float64)
	rFloat, rIsFloat := r.(float64)
	_, lIsString := l.(string)
	_, rIsString := r.(string)

	if lIsFloat && rIsFloat {
		// Accept loss in precision in exchange for simpler code by always using floats for arithmetic.
		switch o {
		case OpMul:
			return lFloat * rFloat, nil
		case OpDiv:
			if rFloat == 0 {
				return nil, errors.New("division by 0")
			}
			return lFloat / rFloat, nil
		case OpAdd:
			return lFloat + rFloat, nil
		case OpSub:
			return lFloat - rFloat, nil
		}
		return nil, errors.New(fmt.Sprintf("unsupported float operator: %v", o))
	}

	if lIsString || rIsString {
		if o == OpAdd {
			return fmt.Sprint(l) + fmt.Sprint(r), nil
		}
		return nil, fmt.Errorf("unsupported string operator (use '+' for concatenation): %v", o)
	}

	return nil, errors.New("unsupported type (only floats and strings are supported)")
}

func (f *Function) eval(ctx Context, caller FunctionCaller) (interface{}, error) {
	var args []interface{}
	for _, arg := range f.Args {
		argEval, err := arg.Value.eval(ctx, caller)
		if err != nil {
			return nil, err
		}
		args = append(args, argEval)
	}
	result, err := caller(f.Name, args...)
	if err != nil {
		return nil, err
	}

	// Convert any int output to float, to simplify parsing.
	resultInt, resultIsInt := result.(int)
	if resultIsInt {
		return float64(resultInt), nil
	}
	return result, nil
}

func (v *Value) eval(ctx Context, caller FunctionCaller) (interface{}, error) {
	switch {
	case v.Number != nil:
		return *v.Number, nil
	case v.StrLiteral != nil:
		return *v.StrLiteral, nil
	case v.Variable != nil:
		value, ok := ctx[*v.Variable]
		if !ok {
			return nil, errors.New("no such variable " + *v.Variable)
		}
		// Attempt to cast to float, then string, then fail.
		valueInt, ok := value.(int)
		if ok {
			return float64(valueInt), nil
		}
		valueFloat, ok := value.(float64)
		if ok {
			return valueFloat, nil
		}
		valueString, ok := value.(string)
		if ok {
			return valueString, nil
		}
		return nil, fmt.Errorf("could not cast variable `%v` to float or string", *v.Variable)
	case v.Function != nil:
		return v.Function.eval(ctx, caller)
	case v.Subexpression != nil:
		return v.Subexpression.eval(ctx, caller)
	default:
		return nil, nil
	}
}

func (f *Factor) eval(ctx Context, caller FunctionCaller) (interface{}, error) {
	b, err := f.Base.eval(ctx, caller)
	if err != nil {
		return nil, err
	}

	if f.Exponent != nil {
		exponentEval, err := f.Exponent.eval(ctx, caller)
		if err != nil {
			return nil, err
		}
		return math.Pow(b.(float64), exponentEval.(float64)), nil
	}
	return b, nil
}

func (t *Term) eval(ctx Context, caller FunctionCaller) (interface{}, error) {
	n, err := t.Left.eval(ctx, caller)
	if err != nil {
		return nil, err
	}

	for _, r := range t.Right {
		rFactorEval, err := r.Factor.eval(ctx, caller)
		if err != nil {
			return nil, err
		}

		n, err = r.Operator.eval(n, rFactorEval)
		if err != nil {
			return nil, err
		}
	}
	return n, nil
}

func (e *Expression) eval(ctx Context, caller FunctionCaller) (interface{}, error) {
	l, err := e.Left.eval(ctx, caller)
	if err != nil {
		return nil, err
	}

	for _, r := range e.Right {
		rEval, err := r.Term.eval(ctx, caller)
		if err != nil {
			return nil, err
		}

		l, err = r.Operator.eval(l, rEval)
		if err != nil {
			return nil, err
		}
	}
	return l, nil
}

// Functions for returning information about expressions.

func (f *Function) identifiers() (variables []string, functions []string) {
	functions = append(functions, f.Name)
	for _, arg := range f.Args {
		argVars, argFuncs := arg.Value.Identifiers()
		variables = append(variables, argVars...)
		functions = append(functions, argFuncs...)
	}
	return variables, functions
}

func (v *Value) identifiers() (variables []string, functions []string) {
	switch {
	case v.Variable != nil:
		variables = append(variables, *v.Variable)
	case v.Function != nil:
		return v.Function.identifiers()
	case v.Subexpression != nil:
		return v.Subexpression.Identifiers()
	}
	return variables, functions
}

func (f *Factor) identifiers() (variables []string, functions []string) {
	variables, functions = f.Base.identifiers()
	if f.Exponent != nil {
		expVars, expFuncs := f.Exponent.identifiers()
		variables = append(variables, expVars...)
		functions = append(functions, expFuncs...)
	}
	return variables, functions
}

func (t *Term) identifiers() (variables []string, functions []string) {
	variables, functions = t.Left.identifiers()
	for _, r := range t.Right {
		rFactorVars, rFactorFuncs := r.Factor.identifiers()
		variables = append(variables, rFactorVars...)
		functions = append(functions, rFactorFuncs...)
	}
	return variables, functions
}

// Identifiers returns the names of the variables and functions in the given expression.
func (e *Expression) Identifiers() (variables []string, functions []string) {
	if e.Left != nil { // Can be nil if the expression is empty (ie: "").
		variables, functions = e.Left.identifiers()
	}
	for _, r := range e.Right {
		opTermVars, opTermFuncs := r.Term.identifiers()
		variables = append(variables, opTermVars...)
		functions = append(functions, opTermFuncs...)
	}
	return variables, functions
}

// Context maps variable names to the values they should be replaced by in expressions.
type Context map[string]interface{}

/*
FunctionCaller defines a function which can call another function given its name as a string and any
arguments.
*/
type FunctionCaller func(string, ...interface{}) (interface{}, error)

/*
Parse is a convenience function which parses a string and returns the resulting expression, which
can then be evaluated.
*/
func Parse(input string) (*Expression, error) {
	expression := &Expression{}
	parser, err := participle.Build(expression)
	if err != nil {
		return nil, fmt.Errorf("could not build parser (try checking the grammar): %v", err)
	}

	if err = parser.ParseString(input, expression); err != nil {
		return nil, fmt.Errorf("could not parse string %q: %v", input, err)
	}
	return expression, nil
}

/*
Eval is a convenience function which evaluates a parsed expression and returns the result.
The ctx parameter is a map containing variable definitions. Note that all numeric variable values
are cast to float64.
*/
func Eval(expression *Expression, ctx Context, caller FunctionCaller) (interface{}, error) {
	result, err := expression.eval(ctx, caller)
	if err != nil {
		return nil, fmt.Errorf("could not evaluate expression `%v`: %v", expression, err)
	}
	glog.Infof("Evaluated expression: %v = %v", expression, result)
	return result, nil
}
