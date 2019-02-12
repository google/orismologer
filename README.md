# Orismologer

_NB: This is not an officially supported Google product_

Orismologer aims to enable clients to get OpenConfig-formatted telemetry from network devices which do not natively support it. This can support scaling of network monitoring infrastructure.

Classically, telemetry is exported via SNMP, CLI scraping or, occasionally, vendor-specific API calls. These methods do not standardise the telemetry format, greatly increasing the complexity and cost of maintaining network monitoring pipelines. The [OpenConfig](openconfig.net) project is addressing this by developing vendor-agnostic data models for network telemetry. However, not all devices support OpenConfig, hence this project. 

Orismologer includes:

1) A protobuf-based framework for expressing translations from 'classic' telemetry to OpenConfig-formatted telemetry (see the `.proto` files in `/proto`). 
2) Specific translations for Cisco and Aruba devices defined using the framework from 1) (see the `.pb` files in `/proto`).
3) Code for evaluating translations defined in the framework from 1) (see the various Go files in this project).
4) A command line interface for 3), supporting commands to a) retrieve data for an OpenConfig path from a hardware target which does not natively support OpenConfig and b) output supported OpenConfig paths (see `oc_translate.go`).

## Install

- Get the code: `go get github.com/google/orismologer`
- Install the [protobuf compiler](https://github.com/protocolbuffers/protobuf). Eg: 
  - Download a [prebuilt binary](https://github.com/protocolbuffers/protobuf/releases)
  - Unzip it
  - Move it onto your `$PATH`: `sudo cp ~/Downloads/bin/protoc /usr/local/bin`
- Install the [Go protobuf tools](https://github.com/golang/protobuf): `go get -u github.com/golang/protobuf/protoc-gen-go`
  - Make sure `protoc-gen-go` is on your `$PATH` (eg: `export PATH=$PATH:~/go/bin`)
- Compile the Orismologer protobuf definitions (the last command needs to be re-run whenever the proto definitions are modified):
 
```
cd orismologer
mkdir proto_out
protoc --go_out=proto_out proto/*.proto
```

## Run
Get data for a supported OpenConfig path for a given hardware target. Vendor information is used to determine which OIDs can be supported. Note that in its current state the project does not retrieve any classic telemetry (but could easily be modified to do so), and the `target` flag is thus not used. See "system design" below for more information.

`go run oc_translate.go get -path /system/state/boot-time -target t -vendor cisco`

Output logs to stderr (NB: the flag must appear before the command).

`go run oc_translate.go -alsologtostderr get -path /system/state/boot-time -target t -vendor cisco`

Print all paths for which at least one translation has been defined.

`go run oc_translate.go print`

Print all paths in an OpenConfig subtree rooted at the given node for which at least one translation has been defined.

`go run oc_translate.go print -root /system`

## Defining New Mappings
New OpenConfig nodes can be added to Orismologer in `proto/mappings.pb` and new transformations can be defined in `proto/transformations.pb`. See below for an overview of these concepts.

## Test
Run the project's tests like you would for any other Go project, eg:

```
cd orismologer
go test ./...
```

## System Overview

Orismologer's telemetry translation framework is implemented as a protobuf schema. This section provides a brief overview of the framework and the code which uses it. Authoritative documentation can be found in comments in the relevant files in this project.

### Transformations

Orismologer allows users to define logic to map from non-OpenConfig paths or "NocPaths" (like SNMP OIDs) to OpenConfig paths. This logic is split into reusable, atomic units called "transformations" (see the example below). Every transformation has an identifier (see the `bind` field) and at least one string expression which contains the translation logic (see below for more information on the expression syntax). The example below defines a transformation with the identifier "example" whose output is always `1000`: 

```
transformations {
  bind: "memory_KB"
  expressions: "1000"
}
```

One transformation can reuse another via its identifier. This example reuses the `memory_KB` transformation:

```
transformations {
  bind: "memory_MB"
  expressions: "memory_KB / 1000"
}
```

Expressions defined in the same transformation are considered equivalent. This is especially useful for normalising equivalent NocPaths across vendors:

```
transformations {
  bind: "memory_MB"
  expressions: "memory_KB / 1000"
  expressions: "memory_B / 1000000"
}
```

Transformations which reference other transformations can only take us so far. Ultimately concrete data has to be retrieved from a NocPath. The example transformation below defines one NocPath for Cisco and one for Aruba, normalises the output to MB, and binds the result to the `memory_MB` identifier. Other transformations can reuse this output without having to concern themselves with vendor-specific differences. Note that `memory_aruba` defines multiple OIDs. As with expressions, multiple OIDs defined in the same NocPath message are considered to produce equivalent output. 

_NB: At this time only SNMP OIDs are supported as NocPaths (but the process for supporting other kinds is straight-forward)._
_NB: The output of all NocPaths is assumed to be of type string. Thus expressions should call `to_X()` on NocPath output, if appropriate._

```
transformations {
  bind: "memory_MB"
  # Both of the following expressions produce equivalent output.
  expressions: "to_int(memory_cisco) / 1000"  # Imagine Cisco's OID outputs in KB.
  expressions: "to_int(memory_aruba) / 1000000"  # Imagine Aruba's OID outputs in B.

  noc_paths {
    bind: "memory_cisco"
    oids: "1.3.6.1.4.1.9.9.168.99.99.99"  # Not a real OID.
  }

  noc_paths {
    bind: "memory_aruba"
    oids: "1.3.6.1.4.1.14823.99.99.99.99"  # Not a real OID.
    oids: "1.1.1.1.1.1.1"  # Imagine that Aruba memory is available from multiple OIDs.
  }
}
```

Thus, Orismologer's transformations form a graph where the nodes represent sets of logically equivalent statements, and the edges represent dependencies amongst them. 

### Mappings

NocPaths (discussed above) are at one edge of Orismologer's transformation graph. At the other is the OpenConfig tree (modeled in this project as nested proto messages). Each node declares a subpath.

```
nodes {
  subpath {path: "/orismologer"}
}
```

A child node's OpenConfig path is formed by prefixing it with its parents subpaths. In the example below the full OpenConfig path of the child node would be `/orismologer/example/memory`.

_NB: The subpath of the top-level ancestor node should start from the OpenConfig root (ie: `/orismologer` not `orismologer`)._ 

```
nodes {
  subpath {path: "/orismologer"}
  children {
    subpath {path: "example/memory"}  # Full path: /orismologer/example/memory
  }
}
```

OpenConfig nodes link to Orismologer's transformation graph by referencing a transformation's identifier (see the `bind` field in the example below). Thus, in the example below, requesting telemetry for the (fake) OpenConfig path `/orismologer/example/memory` would yield the output of the transformation `memory_MB`.

_NB: It generally only makes sense for OpenConfig leaf nodes to link to a transformation_

```
nodes {
  subpath {path: "/orismologer"}
  children {
    subpath {path: "example/memory"}
    bind: "memory_MB"  # Links to a transformation with the identifier "memory_MB". The output of this transformation will be returned for this OpenConfig path.
  }
}
```


### Transformation Evaluation

Given an OpenConfig path and a hardware target, Orismologer can retrieve telemetry data by walking its transformation graph:

- Look up the node indicated by the given path in the OpenConfig tree.
- Find the transformation for that node (ie: its `bind` field).
- Evaluate that transformation:
  - Evaluate each of its expressions in turn, proceeding with the first expression that can be evaluated and skipping the rest:
    - Evaluate each of the variables in the expression, rejecting the entire expression if one variable cannot be evaluated.
    - If a variable links to another transformation, evaluate that transformation (by repeating this process recursively).
    - If a variable links to a NocPath, ensure that it can be evaluated for the given hardware target. If it can, retrieve the requested data and proceed with the next variable in the expression.
    
#### Example

Imagine we have the following mapping and transformation:

```
nodes {
  subpath {path: "/orismologer"}
  children {
    subpath {path: "example/memory"}
    bind: "memory_MB"
  }
}

transformations {
  bind: "memory_MB"
  # Both of the following expressions produce equivalent output.
  expressions: "to_int(memory_cisco) / 1000"
  expressions: "to_int(memory_aruba) / 1000000"

  noc_paths {
    bind: "memory_cisco"
    oids: "1.3.6.1.4.1.9.9.168.99.99.99"
    samples: "1000"
  }

  noc_paths {
    bind: "memory_aruba"
    oids: "1.3.6.1.4.1.14823.99.99.99.99"
    oids: "1.1.1.1.1.1.1"
    samples: "1000000"
  }
}
```

If we request output for `/orismologer/example/memory` we will see the following output, which can be traced to understand the operation of the program:

```
> go run oc_translate.go -alsologtostderr get -target t -path /orismologer/example/memory -vendor aruba
I0212 16:25:04.536684   50360 orismologer.go:123] found transformation "memory_MB" for path "/orismologer/example/memory"
I0212 16:25:04.536856   50360 orismologer.go:139] evaluating transformation "memory_MB" for target "t" of vendor "aruba"
I0212 16:25:04.536864   50360 orismologer.go:179] storing NocPath "memory_cisco" of transformation "memory_MB"
I0212 16:25:04.536871   50360 orismologer.go:179] storing NocPath "memory_aruba" of transformation "memory_MB"
I0212 16:25:04.536878   50360 orismologer.go:143] evaluating expression `to_int(memory_cisco) / 1000`
I0212 16:25:04.537062   50360 orismologer.go:210] evaluating variable "memory_cisco"
I0212 16:25:04.537073   50360 orismologer.go:152] ignoring NocPath "memory_cisco" as it cannot be resolved for vendor "aruba"
I0212 16:25:04.537080   50360 orismologer.go:156] could not evaluate all variables for expression `to_int(memory_cisco) / 1000`, continuing to next expression
I0212 16:25:04.537086   50360 orismologer.go:143] evaluating expression `to_int(memory_aruba) / 1000000`
I0212 16:25:04.537218   50360 orismologer.go:210] evaluating variable "memory_aruba"
I0212 16:25:04.537224   50360 orismologer.go:283] Requesting NocPath "memory_aruba" from target "t"
I0212 16:25:04.537229   50360 orismologer.go:229] evaluated variable "memory_aruba" = 1000000
I0212 16:25:04.537238   50360 functions.go:158] Calling "to_int" with args: 1000000
I0212 16:25:04.537253   50360 parser.go:442] Evaluated expression: to_int(memory_aruba) / 1e+06 = 1
1
```


### Expression Syntax
Expressions are defined in transformation proto messages. They are evaluated at runtime to carry out the operations needed to translate telemetry from one format to another. The expression syntax, by design, very simple. This limits the complexity of the expressions users can write, improving readability and maintainability, and reducing the scope for security exploits. The expression syntax supports the following features:

- Integer literals.
- Float literals.
- String literals.
- Basic arithmetic operators (+, -, *, /, ^).
- Brackets, and a conventional order of operations, eg: `(3 + 7) / 2 = 5`
- String concatenation, eg: `"hello" + "world" = "hello world"`
- Variables.
- Function calls, eg: `my_func(1, "a")`
- Nested expressions (ie: expressions inside expressions), eg: `1 + my_func(2*2, other_func())`

#### Calling Functions
When function calls are encountered in expressions, Orismologer passes the function name (as a string) and any parameters to a function which is responsible for calling an implementation corresponding to that function name. The current implementation only supports calling predefined "library" functions, to reduce scope for security exploits. These are implemented and registered in `functions/functions.go`.
 

## Project Roadmap

- Implement a NocPathResolver which retrieves real data (eg: from a TSDB), instead of relying on hard coded samples.
- Provide a better interface for consumers of OpenConfig telemetry.
- Support OpenConfig list nodes with multiple keys, eg: `node[k1, k2]`
- Support nested OpenConfig list nodes (and, equivalently, nested SNMP tables).
- Add support for CLI-scraped NocPaths.
- Proto validation.
- Support dry runs (for determining if a mapping exists for a given OpenConfig path and hardware target).
- Add a vendor flag to the print subcommand. Only show paths supported for that vendor.
