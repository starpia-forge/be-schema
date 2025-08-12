package beschema

import (
	"testing"
)

func TestMarshalImplicitSchema(t *testing.T) {
	// Test marshaling a simple map[int]any to explicit schema format
	data := map[int]any{
		1: "test1",
		2: "test2",
	}

	// This should fail initially since MarshalImplicitSchema doesn't exist yet
	result, err := MarshalImplicitSchema(data)
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
	result, err := UnmarshalImplicitSchema(data)
	if err != nil {
		t.Fatalf("UnmarshalImplicitSchema failed: %v", err)
	}

	// Verify the result is a proper map[int]any
	if result == nil {
		t.Fatalf("Expected non-nil result")
	}

	// Check that we get back the original data with correct keys
	if result[1] != "test1" {
		t.Errorf("Expected result[1] = 'test1', got %v", result[1])
	}
	if result[2] != "test2" {
		t.Errorf("Expected result[2] = 'test2', got %v", result[2])
	}

	// Verify we have exactly 2 elements
	if len(result) != 2 {
		t.Errorf("Expected 2 elements in result, got %d", len(result))
	}
}

func TestMarshalUnmarshalImplicitSchemaRoundTrip(t *testing.T) {
	// Test complete marshal/unmarshal cycle
	original := ImplicitSchema{
		1: "test1",
		2: "test2",
		3: 42,
		4: true,
	}

	// Marshal to bytes
	data, err := MarshalImplicitSchema(original)
	if err != nil {
		t.Fatalf("MarshalImplicitSchema failed: %v", err)
	}

	// Unmarshal back to map
	result, err := UnmarshalImplicitSchema(data)
	if err != nil {
		t.Fatalf("UnmarshalImplicitSchema failed: %v", err)
	}

	// Verify the round-trip worked correctly
	if len(result) != len(original) {
		t.Errorf("Expected %d elements, got %d", len(original), len(result))
	}

	for key, expectedValue := range original {
		if actualValue, exists := result[key]; !exists {
			t.Errorf("Expected key %d to exist in result", key)
		} else {
			// Handle JSON type conversion: numbers become float64
			switch expected := expectedValue.(type) {
			case int:
				if actual, ok := actualValue.(float64); ok {
					if float64(expected) != actual {
						t.Errorf("Expected result[%d] = %v, got %v", key, float64(expected), actual)
					}
				} else {
					t.Errorf("Expected result[%d] to be float64, got %T", key, actualValue)
				}
			default:
				if actualValue != expectedValue {
					t.Errorf("Expected result[%d] = %v, got %v", key, expectedValue, actualValue)
				}
			}
		}
	}
}
