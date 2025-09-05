package beschema

import (
	"fmt"
	"strings"
)

// Stream represents a structured data stream containing a magic byte and multiple implicit schemas.
type Stream struct {
	MagicByte []byte
	Schemas   []ImplicitSchema
}

// UnmarshalImplicitStream parses a byte slice into a Stream object containing a magic byte and multiple schemas.
// It requires at least three lines in the input data and supports both Windows (\r\n) and Unix (\n) line endings.
// The first line represents the magic byte, and later lines contain schema data in size + JSON pair format.
// Returns a Stream object on success or an error if the input format is invalid or schema unmarshalling fails.
func UnmarshalImplicitStream(data []byte) (*Stream, error) {
	// Convert data to string
	dataStr := string(data)

	// Detect line ending format
	lineEnding := "\r\n"
	lines := strings.Split(dataStr, "\r\n")
	if len(lines) < 3 {
		// Try splitting by \n only (Unix-style line breaks)
		lines = strings.Split(dataStr, "\n")
		lineEnding = "\n"
		if len(lines) < 3 {
			return nil, fmt.Errorf("invalid stream format: expected at least 3 lines")
		}
	}

	// Parse magic byte from the first line
	magicByte := []byte(lines[0])

	// Skip magic byte and empty line, start parsing data pairs from line 2
	var schemas []ImplicitSchema
	i := 2 // Start after the magic byte (line 0) and empty line (line 1)

	for i < len(lines)-1 { // -1 because we need pairs
		// Skip empty lines
		if strings.TrimSpace(lines[i]) == "" {
			i++
			continue
		}

		// Parse size + data pair
		if i+1 < len(lines) {
			sizeData := fmt.Sprintf("%s%s%s%s", lines[i], lineEnding, lines[i+1], lineEnding)
			schema, err := UnmarshalImplicitSchema([]byte(sizeData), true)
			if err != nil {
				return nil, fmt.Errorf("failed to parse schema at line %d: %v", i, err)
			}
			schemas = append(schemas, schema)
			i += 2 // Move to the next pair
		} else {
			break
		}
	}

	return &Stream{
		MagicByte: magicByte,
		Schemas:   schemas,
	}, nil
}

// MarshalImplicitStream serializes a Stream object into
// a formatted byte slice with a magic byte and JSON-encoded schemas.
// It starts with the Stream's magic byte followed by
// an empty line and appends each schema formatted as size and JSON data.
// Returns a byte slice on success or an error if input stream is nil or schema serialization fails.
func MarshalImplicitStream(stream *Stream) ([]byte, error) {
	if stream == nil {
		return nil, fmt.Errorf("stream cannot be nil")
	}

	// Start with magic byte and empty line
	result := fmt.Sprintf("%s\r\n\r\n", string(stream.MagicByte))

	// Marshal each schema and append to the result
	for _, schema := range stream.Schemas {
		schemaData, err := MarshalImplicitSchema(schema, true)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal schema: %v", err)
		}
		result += string(schemaData)
	}

	return []byte(result), nil
}
