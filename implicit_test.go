package beschema

import (
	"testing"
)

func TestMarshalImplicitSchema(t *testing.T) {
	// Test marshaling a simple slice to explicit schema format
	data := ImplicitSchema{"test1", "test2"}

	// This should fail initially since MarshalImplicitSchema doesn't exist yet
	result, err := MarshalImplicitSchema(data, true)
	if err != nil {
		t.Fatalf("MarshalImplicitSchema failed: %v", err)
	}

	// Verify the result is in the expected format: "size\r\nJSON_data\r\n"
	resultStr := string(result)
	if len(resultStr) == 0 {
		t.Fatalf("Expected non-empty result")
	}

	// Should contain the data in JSON array format ordered by keys
	// Expected: ["test1", "test2"] since keys 1,2 should be ordered
	t.Logf("Result: %s", resultStr)
}

func TestUnmarshalImplicitSchema(t *testing.T) {
	// Test unmarshaling explicit schema format back to map[int]any
	// Using the same data format as produced by MarshalImplicitSchema
	data := []byte("19\r\n[\"test1\",\"test2\"]\r\n")

	// This should fail initially since UnmarshalImplicitSchema is not implemented
	result, err := UnmarshalImplicitSchema(data, true)
	if err != nil {
		t.Fatalf("UnmarshalImplicitSchema failed: %v", err)
	}

	// Verify the result is a proper map[int]any
	if result == nil {
		t.Fatalf("Expected non-nil result")
	}

	// Check that we get back the original data with correct indices
	if result[0] != "test1" {
		t.Errorf("Expected result[0] = 'test1', got %v", result[0])
	}
	if result[1] != "test2" {
		t.Errorf("Expected result[1] = 'test2', got %v", result[1])
	}

	// Verify we have exactly 2 elements
	if len(result) != 2 {
		t.Errorf("Expected 2 elements in result, got %d", len(result))
	}
}

func TestMarshalUnmarshalImplicitSchemaRoundTrip(t *testing.T) {
	// Test complete marshal/unmarshal cycle
	original := ImplicitSchema{"test1", "test2", 42, true}

	// Marshal to bytes
	data, err := MarshalImplicitSchema(original, true)
	if err != nil {
		t.Fatalf("MarshalImplicitSchema failed: %v", err)
	}

	// Unmarshal back to map
	result, err := UnmarshalImplicitSchema(data, true)
	if err != nil {
		t.Fatalf("UnmarshalImplicitSchema failed: %v", err)
	}

	// Verify the round-trip worked correctly
	if len(result) != len(original) {
		t.Errorf("Expected %d elements, got %d", len(original), len(result))
	}

	for i, expectedValue := range original {
		if i >= len(result) {
			t.Errorf("Expected index %d to exist in result", i)
			continue
		}
		actualValue := result[i]

		// Handle JSON type conversion: numbers become float64
		switch expected := expectedValue.(type) {
		case int:
			if actual, ok := actualValue.(float64); ok {
				if float64(expected) != actual {
					t.Errorf("Expected result[%d] = %v, got %v", i, float64(expected), actual)
				}
			} else {
				t.Errorf("Expected result[%d] to be float64, got %T", i, actualValue)
			}
		default:
			if actualValue != expectedValue {
				t.Errorf("Expected result[%d] = %v, got %v", i, expectedValue, actualValue)
			}
		}
	}
}

func TestUnmarshalImplicitSchemaWithHeader(t *testing.T) {
	// Test unmarshaling with header (current behavior)
	data := []byte("19\r\n[\"test1\",\"test2\"]\r\n")

	result, err := UnmarshalImplicitSchema(data, true)
	if err != nil {
		t.Fatalf("UnmarshalImplicitSchema with header failed: %v", err)
	}

	if len(result) != 2 {
		t.Errorf("Expected 2 elements, got %d", len(result))
	}
	if result[0] != "test1" {
		t.Errorf("Expected result[0] = 'test1', got %v", result[0])
	}
	if result[1] != "test2" {
		t.Errorf("Expected result[1] = 'test2', got %v", result[1])
	}
}

func TestUnmarshalImplicitSchemaWithoutHeader(t *testing.T) {
	// Test unmarshaling without header (new behavior)
	data := []byte("[[\"test1\",\"test2\",\"[[null]]\",null,null,null,\"test3\"],[\"test4\",1],[\"test5\",2,\"test6\",3]]")

	result, err := UnmarshalImplicitSchema(data, false)
	if err != nil {
		t.Fatalf("UnmarshalImplicitSchema without header failed: %v", err)
	}

	if len(result) != 3 {
		t.Errorf("Expected 3 elements, got %d", len(result))
	}

	// Check first element is an array
	firstElement, ok := result[0].([]interface{})
	if !ok {
		t.Errorf("Expected result[0] to be array, got %T", result[0])
	} else {
		if len(firstElement) != 7 {
			t.Errorf("Expected first element to have 7 items, got %d", len(firstElement))
		}
		if firstElement[0] != "test1" {
			t.Errorf("Expected first element[0] = 'test1', got %v", firstElement[0])
		}
		if firstElement[6] != "test3" {
			t.Errorf("Expected first element[6] = 'test3', got %v", firstElement[6])
		}
	}
}
