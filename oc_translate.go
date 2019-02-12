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
Command oc_translate retrieves telemetry for a given OpenConfig path from a hardware target which
does not natively support OpenConfig.
*/
package main

import (
	"fmt"

	"flag"
	"github.com/google/orismologer/orismologer"
)

const (
	mappingsFile        = "proto/mappings.pb"
	transformationsFile = "proto/transformations.pb"
	vendorOidsFile      = "proto/vendor_oids.pb"
)

var (
	printCommand = flag.NewFlagSet("print", flag.ExitOnError)
	rootFlag     = printCommand.String("root", "root", "print the subtree rooted "+
		"at the given node")

	getCommand = flag.NewFlagSet("get", flag.ExitOnError)
	ocPathFlag = getCommand.String("path", "", "the OpenConfig path to resolve")
	targetFlag = getCommand.String("target", "", "the hardware target for which"+
		"the OpenConfig path should be resolved")
	vendorFlag = getCommand.String("vendor", "", "the vendor of the hardware "+
		"target")
)

func printUsage() {
	fmt.Println(`usage: orismologer <command> [<args>])
	 print    Print an ASCII representation of the tree of OpenConfig nodes which Orismologer can resolve.
	 get      Resolve an OpenConfig path for a given hardware target.`)
}

func main() {
	flag.Usage = printUsage
	flag.Parse()

	o, err := orismologer.NewOrismologer(mappingsFile, transformationsFile, vendorOidsFile)
	if err != nil {
		fmt.Println(err)
		return
	}

	if len(flag.Args()) == 0 {
		fmt.Println("Provide a command")
		printUsage()
		return
	}

	switch flag.Arg(0) {
	case "print":
		printCommand.Parse(flag.Args()[1:])
	case "get":
		getCommand.Parse(flag.Args()[1:])
	default:
		fmt.Printf("Unknown command %q\n", flag.Arg(0))
		printUsage()
	}

	if printCommand.Parsed() {
		o.PrintOcPaths(*rootFlag)
	}

	if getCommand.Parsed() {
		mandatoryArgsPresent := true
		if *ocPathFlag == "" {
			fmt.Println("supply an OpenConfig path")
			mandatoryArgsPresent = false
		}

		if *targetFlag == "" {
			fmt.Println("supply a hardware target")
			mandatoryArgsPresent = false
		}

		if *vendorFlag == "" {
			fmt.Println("supply the vendor of the hardware target")
			mandatoryArgsPresent = false
		}

		if mandatoryArgsPresent {
			result, err := o.Eval(*ocPathFlag, *targetFlag, *vendorFlag)
			if err != nil {
				fmt.Println(err)
				return
			}
			fmt.Println(result)
		}
	}
}
