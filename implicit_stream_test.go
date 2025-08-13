package beschema

import (
	"os"
	"strings"
	"testing"
)

func TestParseStreamMagicByte(t *testing.T) {
	// Test parsing magic byte from stream data
	streamData := []byte(")]}'\r\n\r\n10\r\n[\"test\"]\r\n")

	stream, err := UnmarshalImplicitStream(streamData)
	if err != nil {
		t.Fatalf("UnmarshalImplicitStream failed: %v", err)
	}

	expectedMagicByte := []byte(")]}'")
	if string(stream.MagicByte) != string(expectedMagicByte) {
		t.Errorf("Expected magic byte %q, got %q", string(expectedMagicByte), string(stream.MagicByte))
	}
}

func TestParseEmptyStream(t *testing.T) {
	// Test parsing empty stream (magic byte + empty lines only)
	streamData := []byte(")]}'\r\n\r\n")

	stream, err := UnmarshalImplicitStream(streamData)
	if err != nil {
		t.Fatalf("UnmarshalImplicitStream failed: %v", err)
	}

	expectedMagicByte := []byte(")]}'")
	if string(stream.MagicByte) != string(expectedMagicByte) {
		t.Errorf("Expected magic byte %q, got %q", string(expectedMagicByte), string(stream.MagicByte))
	}

	if len(stream.Schemas) != 0 {
		t.Errorf("Expected empty schemas, got %d schemas", len(stream.Schemas))
	}
}

func TestParseStreamWithSingleSchema(t *testing.T) {
	// Test parsing stream with single schema entry
	streamData := []byte(")]}'\r\n\r\n19\r\n[\"test1\",\"test2\"]\r\n")

	stream, err := UnmarshalImplicitStream(streamData)
	if err != nil {
		t.Fatalf("UnmarshalImplicitStream failed: %v", err)
	}

	expectedMagicByte := []byte(")]}'")
	if string(stream.MagicByte) != string(expectedMagicByte) {
		t.Errorf("Expected magic byte %q, got %q", string(expectedMagicByte), string(stream.MagicByte))
	}

	if len(stream.Schemas) != 1 {
		t.Fatalf("Expected 1 schema, got %d schemas", len(stream.Schemas))
	}

	schema := stream.Schemas[0]
	if schema[0] != "test1" {
		t.Errorf("Expected schema[0] = 'test1', got %v", schema[0])
	}
	if schema[1] != "test2" {
		t.Errorf("Expected schema[1] = 'test2', got %v", schema[1])
	}
}

func TestParseStreamWithMultipleSchemas(t *testing.T) {
	// Test parsing stream with multiple schema entries
	streamData := []byte(")]}'\r\n\r\n19\r\n[\"test1\",\"test2\"]\r\n14\r\n[\"data1\",42]\r\n")

	stream, err := UnmarshalImplicitStream(streamData)
	if err != nil {
		t.Fatalf("UnmarshalImplicitStream failed: %v", err)
	}

	expectedMagicByte := []byte(")]}'")
	if string(stream.MagicByte) != string(expectedMagicByte) {
		t.Errorf("Expected magic byte %q, got %q", string(expectedMagicByte), string(stream.MagicByte))
	}

	if len(stream.Schemas) != 2 {
		t.Fatalf("Expected 2 schemas, got %d schemas", len(stream.Schemas))
	}

	// Check first schema
	schema1 := stream.Schemas[0]
	if schema1[0] != "test1" {
		t.Errorf("Expected schema1[0] = 'test1', got %v", schema1[0])
	}
	if schema1[1] != "test2" {
		t.Errorf("Expected schema1[1] = 'test2', got %v", schema1[1])
	}

	// Check second schema
	schema2 := stream.Schemas[1]
	if schema2[0] != "data1" {
		t.Errorf("Expected schema2[0] = 'data1', got %v", schema2[0])
	}
	// JSON unmarshaling converts numbers to float64
	if schema2[1] != float64(42) {
		t.Errorf("Expected schema2[1] = 42.0, got %v", schema2[1])
	}
}

func TestHandleInvalidMagicByteFormat(t *testing.T) {
	// Test handling invalid magic byte format - data too short
	streamData := []byte("x\r\n")

	_, err := UnmarshalImplicitStream(streamData)
	if err == nil {
		t.Fatalf("Expected error for invalid stream format, got nil")
	}

	expectedError := "invalid stream format: expected at least 3 lines"
	if err.Error() != expectedError {
		t.Errorf("Expected error %q, got %q", expectedError, err.Error())
	}
}

func TestHandleMissingLineBreaksAfterMagicByte(t *testing.T) {
	// Test handling missing line breaks after magic byte - no empty line after magic byte
	// This should be treated as invalid format since the spec requires 2 line breaks after magic byte
	streamData := []byte(")]}'\r\n10\r\n[\"test\"]\r\n")

	_, err := UnmarshalImplicitStream(streamData)
	if err == nil {
		t.Fatalf("Expected error for missing line breaks after magic byte, got nil")
	}

	// Should get an error about invalid size format since "10" is at wrong position
	expectedErrorSubstring := "invalid size format"
	if !strings.Contains(err.Error(), expectedErrorSubstring) {
		t.Errorf("Expected error containing %q, got %q", expectedErrorSubstring, err.Error())
	}
}

func TestHandleInvalidSizeFormatInDataPairs(t *testing.T) {
	// Test handling invalid size format in data pairs - non-numeric size
	streamData := []byte(")]}'\r\n\r\nabc\r\n[\"test\"]\r\n")

	_, err := UnmarshalImplicitStream(streamData)
	if err == nil {
		t.Fatalf("Expected error for invalid size format, got nil")
	}

	// Should get an error about invalid size format
	expectedErrorSubstring := "invalid size format"
	if !strings.Contains(err.Error(), expectedErrorSubstring) {
		t.Errorf("Expected error containing %q, got %q", expectedErrorSubstring, err.Error())
	}
}

func TestHandleSizeMismatchInDataPairs(t *testing.T) {
	// Test handling size mismatch in data pairs - declared size doesn't match actual data size
	streamData := []byte(")]}'\r\n\r\n20\r\n[\"test\"]\r\n")

	_, err := UnmarshalImplicitStream(streamData)
	if err == nil {
		t.Fatalf("Expected error for size mismatch, got nil")
	}

	// Should get an error about data size mismatch
	expectedErrorSubstring := "data size mismatch"
	if !strings.Contains(err.Error(), expectedErrorSubstring) {
		t.Errorf("Expected error containing %q, got %q", expectedErrorSubstring, err.Error())
	}
}

func TestHandleInvalidJSONInDataPairs(t *testing.T) {
	// Test handling invalid JSON in data pairs - malformed JSON
	streamData := []byte(")]}'\r\n\r\n18\r\n[\"test\",invalid]\r\n")

	_, err := UnmarshalImplicitStream(streamData)
	if err == nil {
		t.Fatalf("Expected error for invalid JSON, got nil")
	}

	// Should get an error about failed to unmarshal JSON
	expectedErrorSubstring := "failed to unmarshal JSON"
	if !strings.Contains(err.Error(), expectedErrorSubstring) {
		t.Errorf("Expected error containing %q, got %q", expectedErrorSubstring, err.Error())
	}
}

func TestMarshalStreamToByteArray(t *testing.T) {
	// Test marshaling Stream to byte array with magic byte and schemas
	stream := &Stream{
		MagicByte: []byte(")]}'"),
		Schemas: []ImplicitSchema{
			{"test1", "test2"},
			{"data1", 42},
		},
	}

	result, err := MarshalImplicitStream(stream)
	if err != nil {
		t.Fatalf("MarshalImplicitStream failed: %v", err)
	}

	// Verify the result starts with magic byte and empty line
	resultStr := string(result)
	if !strings.HasPrefix(resultStr, ")]}'\r\n\r\n") {
		t.Errorf("Expected result to start with magic byte and empty line, got: %s", resultStr[:20])
	}

	// Verify we can unmarshal it back
	unmarshaled, err := UnmarshalImplicitStream(result)
	if err != nil {
		t.Fatalf("Failed to unmarshal result: %v", err)
	}

	// Verify magic byte matches
	if string(unmarshaled.MagicByte) != string(stream.MagicByte) {
		t.Errorf("Expected magic byte %q, got %q", string(stream.MagicByte), string(unmarshaled.MagicByte))
	}

	// Verify schema count matches
	if len(unmarshaled.Schemas) != len(stream.Schemas) {
		t.Errorf("Expected %d schemas, got %d", len(stream.Schemas), len(unmarshaled.Schemas))
	}
}

func TestMarshalEmptyStream(t *testing.T) {
	// Test marshaling empty Stream (magic byte only)
	stream := &Stream{
		MagicByte: []byte(")]}'"),
		Schemas:   []ImplicitSchema{},
	}

	result, err := MarshalImplicitStream(stream)
	if err != nil {
		t.Fatalf("MarshalImplicitStream failed: %v", err)
	}

	// Should only contain magic byte and empty line
	expected := ")]}'\r\n\r\n"
	if string(result) != expected {
		t.Errorf("Expected %q, got %q", expected, string(result))
	}

	// Verify we can unmarshal it back
	unmarshaled, err := UnmarshalImplicitStream(result)
	if err != nil {
		t.Fatalf("Failed to unmarshal result: %v", err)
	}

	// Verify magic byte matches
	if string(unmarshaled.MagicByte) != string(stream.MagicByte) {
		t.Errorf("Expected magic byte %q, got %q", string(stream.MagicByte), string(unmarshaled.MagicByte))
	}

	// Verify no schemas
	if len(unmarshaled.Schemas) != 0 {
		t.Errorf("Expected 0 schemas, got %d", len(unmarshaled.Schemas))
	}
}

func TestParseCompleteStreamFile(t *testing.T) {
	// Test parsing complete sample stream.txt file
	streamData, err := os.ReadFile("test\\sample\\stream.txt")
	if err != nil {
		t.Fatalf("Failed to read stream.txt: %v", err)
	}

	stream, err := UnmarshalImplicitStream(streamData)
	if err != nil {
		t.Fatalf("UnmarshalImplicitStream failed: %v", err)
	}

	expectedMagicByte := []byte(")]}'")
	if string(stream.MagicByte) != string(expectedMagicByte) {
		t.Errorf("Expected magic byte %q, got %q", string(expectedMagicByte), string(stream.MagicByte))
	}

	// The sample file should have 6 schema entries based on the data pairs
	expectedSchemaCount := 6
	if len(stream.Schemas) != expectedSchemaCount {
		t.Errorf("Expected %d schemas, got %d schemas", expectedSchemaCount, len(stream.Schemas))
	}

	// Verify each schema was parsed successfully
	for i, schema := range stream.Schemas {
		if len(schema) == 0 {
			t.Errorf("Schema %d is empty", i)
		}
	}
}
