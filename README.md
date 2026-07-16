# Easy JSON

A Go package for ease and simple operations with JSON data.

![Lint][Lint-Status-Image-Url]

[Lint-Status-Image-Url]: https://github.com/foliagecp/easyjson/actions/workflows/golangci-lint.yml/badge.svg

## Overview

Easy JSON is a Go package that provides a convenient way to work with JSON data. It offers path-based access, a fluent builder API, deep merging, normalization to a canonical form, and typed value extraction — all with zero external dependencies.

## Installation

To use Easy JSON in your Go project, you can simply import it using Go modules:

```go
import "github.com/foliagecp/easyjson"
```

## Usage

### Creating JSON Values

```go
obj := easyjson.NewJSONObject()                                  // {}
arr := easyjson.NewJSONArray()                                   // []
null := easyjson.NewJSONNull()                                   // null
kv := easyjson.NewJSONObjectWithKeyValue("id", easyjson.NewJSON(1)) // {"id":1}
```

`NewJSON` accepts any Go value — primitives, slices, maps, and structs (structs are converted via JSON marshaling):

```go
j := easyjson.NewJSON([]string{"a", "b"})            // ["a","b"]
j = easyjson.NewJSON(map[string]int{"x": 1})         // {"x":1}
j = easyjson.NewJSON(struct{ Name string }{"John"})  // {"Name":"John"}
```

> `JSONFromArray` is deprecated — use `NewJSON` instead, it handles slices properly.

`NewJSONBytes` stores a `[]byte` as a hex-encoded string; `AsBytes` decodes it back:

```go
j := easyjson.NewJSONBytes([]byte{0xDE, 0xAD})
b, ok := j.AsBytes() // []byte{0xDE, 0xAD}, true
```

### Parsing JSON

```go
j, ok := easyjson.JSONFromString(`{"user": {"name": "John"}}`)
j, ok = easyjson.JSONFromBytes([]byte(`[1, 2, 3]`))
```

### Path Operations

Paths are dot-separated by default; numeric tokens index into arrays.

```go
obj.SetByPath("user.name", easyjson.NewJSON("John"))
name := obj.GetByPath("user.name").ToString()
exists := obj.PathExists("user.name")
obj.RemoveByPath("user.name")
```

Notes on `SetByPath` semantics:
- Missing intermediate nodes are created as **objects** — arrays are never created implicitly, even for a numeric path token. Create them explicitly with `NewJSONArray()` first.
- A negative index as the last path token appends to an existing array: `obj.SetByPath("list.-1", v)`.
- `RemoveByPath` on an array element sets it to `null` rather than shrinking the array.

All path functions accept an optional custom delimiter:

```go
obj.SetByPath("user/name", easyjson.NewJSON("John"), "/")
name := obj.GetByPath("user/name", "/").ToString()
```

### Setting Multiple Paths at Once

```go
obj.SetByPaths(map[string]interface{}{
    "user.name": "John",
    "user.age":  30,
})

obj = easyjson.NewJSONObjectFromMap(map[string]interface{}{
    "ip_address": "10.0.0.1",
    "port":       8080,
})
```

### Builder API

`JSONBuilder` provides a fluent interface for constructing JSON:

```go
j := easyjson.NewJSONBuilder().
    Set("user.name", "John").
    SetIfNotEmpty("user.email", email). // skipped if nil, "" or 0
    AddToArray("user.roles", "admin").
    Build()
```

`BuildArrayFromSlice` maps a Go slice into a JSON array of objects:

```go
arr := easyjson.BuildArrayFromSlice(users, func(u User) map[string]interface{} {
    return map[string]interface{}{"name": u.Name, "age": u.Age}
})
```

### Type Checks and Value Extraction

```go
j.IsObject(); j.IsArray(); j.IsString(); j.IsNumeric(); j.IsBool(); j.IsNull()
j.IsNonEmptyObject(); j.IsNonEmptyArray()

s, ok := j.AsString()
f, ok := j.AsNumeric()      // float64
i, ok := j.AsInt64()        // int64 without precision loss through float64
b, ok := j.AsBool()
m, ok := j.AsObject()       // map[string]interface{}
a, ok := j.AsArray()        // []interface{}
ss, ok := j.AsArrayString() // []string, if all elements are strings
```

Each extractor has a `...Default` variant returning a fallback instead of a bool:

```go
port := j.GetByPath("config.port").AsInt64Default(8080)
name := j.GetByPath("user.name").AsStringDefault("anonymous")
```

### Arrays and Objects

```go
arr.AddToArray(easyjson.NewJSON("x"))
n := arr.ArraySize()          // -1 if not an array
el := arr.ArrayElement(0)

keys := obj.ObjectKeys()
count := obj.KeysCount()
```

### Deep Merging

```go
obj1.DeepMerge(obj2) // recursively merges obj2 into obj1
```

### Comparing and Normalizing

`Equals` performs deep equality (`reflect.DeepEqual`). For order-insensitive comparison, normalize first — `Normalize` converts all numbers to `float64` and sorts arrays deterministically by the canonical string of their elements:

```go
isEqual := obj1.Equals(obj2)
isEquivalent := obj1.NormalizedClone().Equals(obj2.NormalizedClone())

obj.Normalize() // in place
```

### Cloning and Serialization

```go
copy := j.Clone()   // deep copy
s := j.ToString()
b := j.ToBytes()
```

## Documentation

For more details and usage examples, please refer to the [official documentation](https://pkg.go.dev/github.com/foliagecp/easyjson).

## License

Unless otherwise noted, the easyjson source files are distributed under the Apache Version 2.0 license found in the LICENSE file.

## Contribution

Contributions and bug reports are welcome! Please submit issues or pull requests to help improve this package.
