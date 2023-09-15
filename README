# Easy JSON

A Go package for ease and simple operations with JSON data.

## Overview

Easy JSON is a Go package that provides a convenient way to work with JSON data. It offers functionality for creating, manipulating, and comparing JSON objects and arrays.

## Installation

To use Easy JSON in your Go project, you can simply import it using Go modules:

```go
import "github.com/foliagecp/easyjson"
```

## Usage

### Creating a JSON Object

```go
obj := easyjson.NewJSONObject()
```

### Creating a JSON Array

```go
arr := easyjson.NewJSONArray()
```

### Setting Values by Path

```go
obj.SetByPath("user.name", easyjson.NewJSON("John"))
```

### Getting Values by Path

```go
name := obj.GetByPath("user.name").ToString()
```

### Deep Merging JSON Objects

```go
obj1 := easyjson.NewJSONObject()
obj2 := easyjson.NewJSONObject()

obj1.DeepMerge(obj2)
```

### Comparing JSON Objects

```go
isEqual := obj1.Equals(obj2)
```

## Documentation

For more details and usage examples, please refer to the [official documentation](https://pkg.go.dev/github.com/foliagecp/easyjson).

## License

Unless otherwise noted, the easyjson source files are distributed under the Apache Version 2.0 license found in the LICENSE file.

## Contribution

Contributions and bug reports are welcome! Please submit issues or pull requests to help improve this package.