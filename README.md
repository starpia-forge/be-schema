## Languages | 언어

- [English](README.md) (Current)
- [한국어](README_ko.md)

---

# be-schema

This library is for serializing or deserializing API Request, Response Payload using `Google Batchexecute`.

Although I don't know which fields `Google` actually uses, I created it so that library users can infer and parse fields using arbitrary explicit struct fields.

## Features

- Convert Go structs to ordered arrays based on `beschema` tags
- Convert arrays back to structs with proper field mapping
- Support for nested structs
- Explicit field ordering control
- JSON marshaling/unmarshaling with schema-based ordering

## Installation

```bash
go get github.com/starpia-forge/be-schema
```

## Usage

### Basic Example

```go
package main

import (
    "fmt"
    "github.com/starpia-forge/be-schema"
)

type Person struct {
    Name string `beschema:"1"`
    Age  int    `beschema:"2"`
    City string `beschema:"3"`
}

func main() {
    person := Person{
        Name: "John Doe",
        Age:  30,
        City: "New York",
    }

    // Marshal to JSON with explicit schema ordering
    data, err := beschema.MarshalExplicitSchema(person)
    if err != nil {
        panic(err)
    }
    fmt.Printf("JSON: %s\n", data)

    // Unmarshal back to struct
    var result Person
    result, err = beschema.UnmarshalExplicitSchema[Person](data)
    if err != nil {
        panic(err)
    }
    fmt.Printf("Struct: %+v\n", result)
}
```

### Nested Structs

```go
type Address struct {
    Street string `beschema:"1"`
    City   string `beschema:"2"`
}

type Person struct {
    Name    string  `beschema:"1"`
    Address Address `beschema:"2"`
    Age     int     `beschema:"3"`
}

func main() {
    person := Person{
        Name: "Jane Doe",
        Address: Address{
            Street: "123 Main St",
            City:   "Boston",
        },
        Age: 25,
    }

    data, err := beschema.MarshalExplicitSchema(person)
    if err != nil {
        panic(err)
    }
    
    // Output: [["Jane Doe", ["123 Main St", "Boston"], 25]]
    fmt.Printf("JSON: %s\n", data)
}
```

## API Reference

### Functions

#### `MarshalExplicitSchema[T any](v T) ([]byte, error)`

Marshals a struct to JSON array format using `beschema` tag ordering.

**Parameters:**
- `v`: The struct to marshal

**Returns:**
- `[]byte`: JSON-encoded array
- `error`: Error if marshaling fails

#### `UnmarshalExplicitSchema[T any](data []byte) (T, error)`

Unmarshals JSON array data back to a struct using `beschema` tag ordering.

**Parameters:**
- `data`: JSON-encoded array data

**Returns:**
- `T`: The unmarshaled struct
- `error`: Error if unmarshaling fails

## Schema Tags

Use the `beschema` tag to specify the order of fields in the resulting array:

```go
type Example struct {
    Third  string `beschema:"3"`
    First  string `beschema:"1"`
    Second string `beschema:"2"`
}
```

Fields will be ordered in the array as: `[First, Second, Third]` regardless of their declaration order in the struct.

## Requirements

- Go 1.24 or later

## License

This project is licensed under the [MIT License](LICENSE).