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
Package functions maps a collection of functions to string keys and facilitates calling them with
these keys.
Registered functions must return 1 or 2 values. If 2, then the second must be an error (or nil).
*/
package functions

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/golang/glog"
	"github.com/google/orismologer/utils"
)

// Functions must be registered here to expose them to external callers.
var registry = map[string]interface{}{
	"to_int":           toInt,
	"to_str":           toStr,
	"time_since_epoch": timeSinceEpoch,
}

// Implementations of functions.

func toStr(value interface{}) (string, error) {
	result, ok := value.(string)
	if !ok {
		return "", fmt.Errorf("value `%v` could not be cast to string", value)
	}
	return result, nil
}

func toInt(value interface{}) (int, error) {
	if str, err := toStr(value); err == nil {
		if result, err := strconv.Atoi(str); err == nil {
			return result, nil
		}
	}
	result, ok := value.(int)
	if !ok {
		return 0, fmt.Errorf("value `%v` could not be cast to int", value)
	}
	return result, nil
}

func toFloat(value interface{}) (float64, error) {
	if str, err := toStr(value); err == nil {
		if result, err := strconv.ParseFloat(str, 64); err == nil {
			return result, nil
		}
	}
	result, ok := value.(float64)
	if !ok {
		return 0, fmt.Errorf("value `%v` could not be cast to float64", value)
	}
	return result, nil
}

/*
timeSinceEpoch returns the amount of time since the Unix epoch (1970-01-01) in the requested units.
Format can be "rfc3339", "ntp" or any time format string understood by Go's time.Parse().
Units can be "s", "ms" or "ns".
*/
func timeSinceEpoch(value interface{}, format string, units string) (int, error) {
	timeStamp, ok := value.(string)
	if !ok {
		return 0, fmt.Errorf("requested %v to unix conversion, but %q is not %v formatted", format, value, format)
	}
	var t time.Time
	switch format {
	case "ntp":
		timeStamp = strings.Replace(timeStamp, " ", "", -1)
		const offset = 2208988800 // NTP epoch is 1900-01-01.
		ntp, err := strconv.ParseUint(timeStamp, 16, 64)
		if err != nil {
			fmt.Println(err)
		}
		seconds := ntp>>32 - offset
		fractional := ntp & 0x00000000ffffffff
		nanos := fractional * 1000000000 >> 32
		t = time.Unix(int64(seconds), int64(nanos))
	case "rfc3339":
		format = time.RFC3339
		fallthrough
	default:
		var err error
		t, err = time.Parse(format, timeStamp)
		if err != nil {
			return 0, fmt.Errorf("error parsing timestamp %q of format %q: %v", value, format, err)
		}
	}
	switch units {
	case "s":
		return int(t.Unix()), nil
	case "ms":
		return int(t.UnixNano() / 1000000), nil
	case "ns":
		return int(t.UnixNano()), nil
	default:
		return 0, fmt.Errorf("unrecognised unit %q", units)
	}
}

// Code to handle and call library functions.

/*
Library contains a predefined collection of functions which may be called via a string key.
*/
type Library struct {
	functions map[string]interface{}
}

// NewLibrary returns a new function library.
func NewLibrary() Library {
	return newLibrary(registry)
}

func newLibrary(registry map[string]interface{}) Library {
	return Library{functions: registry}
}

/*
Call calls a function from a predefined collected, given only the function's name as a string and
any arguments to be passed to it.
*/
func (l Library) Call(funcName string, args ...interface{}) (interface{}, error) {
	f, err := l.getFunc(funcName)
	if err != nil {
		return nil, err
	}

	numArgsExpected := f.Type().NumIn()
	numArgs := len(args)
	if numArgs != numArgsExpected {
		return nil, fmt.Errorf("function %q expects %v arguments, but got %v", funcName, numArgsExpected, numArgs)
	}

	wrappedArgs := wrapArgs(args...)
	glog.Info(fmt.Sprintf("Calling %q with args: %v\n", funcName, utils.SliceToString(args)))
	output := f.Call(wrappedArgs)
	return unwrapOutput(output, funcName)
}

func (l Library) getFunc(funcName string) (reflect.Value, error) {
	if !l.Contains(funcName) {
		return reflect.Value{}, fmt.Errorf("function %q undefined", funcName)
	}
	return reflect.ValueOf(l.functions[funcName]), nil
}

// wrapArgs wraps each arg in a reflect.Value.
func wrapArgs(args ...interface{}) []reflect.Value {
	wrappedArgs := make([]reflect.Value, len(args))
	for i, arg := range args {
		wrappedArgs[i] = reflect.ValueOf(arg)
	}
	return wrappedArgs
}

// unwrapOutput unwraps output wrapped in reflect.Value.
func unwrapOutput(output []reflect.Value, funcName string) (interface{}, error) {
	results := make([]interface{}, len(output))
	for i, value := range output {
		results[i] = value.Interface()
	}
	switch len(results) {
	case 1:
		return results[0], nil
	case 2:
		result, wrappedErr := results[0], results[1]
		if wrappedErr == nil {
			return result, nil
		}
		err, ok := wrappedErr.(error)
		if !ok {
			return nil, fmt.Errorf("function %q returned two values, but the second was not an error. The value was: %v", funcName, wrappedErr)
		}
		return result, err
	default:
		return nil, fmt.Errorf("function %q returned unexpected number of results", funcName)
	}
}

// Contains returns true if a function with the given name has been defined.
func (l Library) Contains(funcName string) bool {
	return l.functions[funcName] != nil
}
