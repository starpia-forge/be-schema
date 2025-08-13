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

### Examples

This library provides three different ways to handle schema marshaling/unmarshaling:

#### 1. Explicit Schema
For structured data with predefined struct types and `beschema` tags:
- See: [`cmd/example/explicit/explicit.go`](cmd/example/explicit/explicit.go)

#### 2. Implicit Schema  
For dynamic data without predefined struct types:
- See: [`cmd/example/implicit/implicit.go`](cmd/example/implicit/implicit.go)

#### 3. Implicit Stream
For streaming data with implicit schema:
- See: [`cmd/example/implicit_stream/implicit_stream.go`](cmd/example/implicit_stream/implicit_stream.go)

### Running Examples

You can run any of the examples using:

```bash
# Explicit schema example
go run cmd/example/explicit/explicit.go

# Implicit schema example  
go run cmd/example/implicit/implicit.go

# Implicit stream example
go run cmd/example/implicit_stream/implicit_stream.go
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