package beschema

import (
	"encoding/json"
	"strings"
	"testing"
)

// Test struct with beschema tags - field order different from tag order
type TestStruct struct {
	Field2 string `beschema:"2"`
	Field1 string `beschema:"1"`
}

func TestStructToArrayWithBeschemaTag(t *testing.T) {
	// Test data with fields in different order than struct definition
	testData := TestStruct{
		Field1: "first",
		Field2: "second",
	}

	// Expected array should respect beschema tag order: [Field1, Field2]
	expected := []interface{}{"first", "second"}

	result, err := structToArray(testData)
	if err != nil {
		t.Fatalf("structToArray failed: %v", err)
	}

	if len(result) != len(expected) {
		t.Fatalf("Expected array length %d, got %d", len(expected), len(result))
	}

	for i, expectedVal := range expected {
		if result[i] != expectedVal {
			t.Errorf("Expected result[%d] = %v, got %v", i, expectedVal, result[i])
		}
	}
}

func TestArrayToStructWithBeschemaTag(t *testing.T) {
	// Array data ordered by beschema tags: [Field1, Field2]
	arrayData := []interface{}{"first", "second"}

	var result TestStruct
	err := arrayToStruct(arrayData, &result)
	if err != nil {
		t.Fatalf("arrayToStruct failed: %v", err)
	}

	// Verify fields are populated correctly according to beschema tags
	if result.Field1 != "first" {
		t.Errorf("Expected Field1 = 'first', got '%s'", result.Field1)
	}
	if result.Field2 != "second" {
		t.Errorf("Expected Field2 = 'second', got '%s'", result.Field2)
	}
}

// Test nested structs with beschema tags
type NestedStruct struct {
	InnerField2 string `beschema:"2"`
	InnerField1 string `beschema:"1"`
}

type OuterStruct struct {
	Nested NestedStruct `beschema:"1"`
	Field  string       `beschema:"2"`
}

func TestNestedStructWithBeschemaTag(t *testing.T) {
	// Test data with nested struct
	testData := OuterStruct{
		Nested: NestedStruct{
			InnerField1: "inner1",
			InnerField2: "inner2",
		},
		Field: "outer",
	}

	// Expected array should respect beschema tag order: [Nested, Field]
	// Nested should be ordered by its tags: [InnerField1, InnerField2]
	expected := []interface{}{[]interface{}{"inner1", "inner2"}, "outer"}

	result, err := structToArray(testData)
	if err != nil {
		t.Fatalf("structToArray failed: %v", err)
	}

	if len(result) != len(expected) {
		t.Fatalf("Expected array length %d, got %d", len(expected), len(result))
	}

	// Check nested array
	if nestedArr, ok := result[0].([]interface{}); ok {
		expectedNested := expected[0].([]interface{})
		if len(nestedArr) != len(expectedNested) {
			t.Fatalf("Expected nested array length %d, got %d", len(expectedNested), len(nestedArr))
		}
		for i, expectedVal := range expectedNested {
			if nestedArr[i] != expectedVal {
				t.Errorf("Expected nested[%d] = %v, got %v", i, expectedVal, nestedArr[i])
			}
		}
	} else {
		t.Errorf("Expected nested array, got %T", result[0])
	}

	// Check outer field
	if result[1] != expected[1] {
		t.Errorf("Expected result[1] = %v, got %v", expected[1], result[1])
	}
}

func TestArrayToNestedStructWithBeschemaTag(t *testing.T) {
	// Array data with a nested array: [Nested, Field]
	// Nested array ordered by beschema tags: [InnerField1, InnerField2]
	arrayData := []interface{}{[]interface{}{"inner1", "inner2"}, "outer"}

	var result OuterStruct
	err := arrayToStruct(arrayData, &result)
	if err != nil {
		t.Fatalf("arrayToStruct failed: %v", err)
	}

	// Verify outer field
	if result.Field != "outer" {
		t.Errorf("Expected Field = 'outer', got '%s'", result.Field)
	}

	// Verify nested struct fields are populated correctly according to beschema tags
	if result.Nested.InnerField1 != "inner1" {
		t.Errorf("Expected InnerField1 = 'inner1', got '%s'", result.Nested.InnerField1)
	}
	if result.Nested.InnerField2 != "inner2" {
		t.Errorf("Expected InnerField2 = 'inner2', got '%s'", result.Nested.InnerField2)
	}
}

func TestMarshalUnmarshalExplicitSchemaWithBeschemaTag(t *testing.T) {
	// Test complete marshal/unmarshal cycle with beschema tags
	original := OuterStruct{
		Nested: NestedStruct{
			InnerField1: "inner1",
			InnerField2: "inner2",
		},
		Field: "outer",
	}

	// Marshal to bytes
	data, err := MarshalExplicitSchema(original)
	if err != nil {
		t.Fatalf("MarshalExplicitSchema failed: %v", err)
	}

	// Unmarshal back to struct
	result, err := UnmarshalExplicitSchema[OuterStruct](data)
	if err != nil {
		t.Fatalf("UnmarshalExplicitSchema failed: %v", err)
	}

	// Verify the round-trip worked correctly
	if result.Field != original.Field {
		t.Errorf("Expected Field = '%s', got '%s'", original.Field, result.Field)
	}
	if result.Nested.InnerField1 != original.Nested.InnerField1 {
		t.Errorf("Expected InnerField1 = '%s', got '%s'", original.Nested.InnerField1, result.Nested.InnerField1)
	}
	if result.Nested.InnerField2 != original.Nested.InnerField2 {
		t.Errorf("Expected InnerField2 = '%s', got '%s'", original.Nested.InnerField2, result.Nested.InnerField2)
	}

	// Verify the data format matches the expected structure
	dataStr := string(data)
	lines := strings.Split(dataStr, "\r\n")
	if len(lines) < 2 {
		t.Fatalf("Expected at least 2 lines in output, got %d", len(lines))
	}

	// Parse the JSON to verify structure
	var arr []interface{}
	if err := json.Unmarshal([]byte(lines[1]), &arr); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	// Should be [Nested, Field] where Nested is [InnerField1, InnerField2]
	if len(arr) != 2 {
		t.Fatalf("Expected 2 elements in array, got %d", len(arr))
	}

	// Check nested array
	if nestedArr, ok := arr[0].([]interface{}); ok {
		if len(nestedArr) != 2 {
			t.Fatalf("Expected 2 elements in nested array, got %d", len(nestedArr))
		}
		if nestedArr[0] != "inner1" || nestedArr[1] != "inner2" {
			t.Errorf("Expected nested array ['inner1', 'inner2'], got %v", nestedArr)
		}
	} else {
		t.Errorf("Expected nested array, got %T", arr[0])
	}

	// Check the outer field
	if arr[1] != "outer" {
		t.Errorf("Expected second element 'outer', got %v", arr[1])
	}
}

// Test structs that reproduce the exact issue from cmd/example/explicit/explicit.go
type Entity struct {
	Sub1 SubEntity1 `beschema:"1"`
	Sub2 SubEntity2 `beschema:"3"`
}

type SubEntity1 struct {
	Field1 string `beschema:"1"`
	Field2 string `beschema:"2"`
}

type SubEntity2 struct {
	Field1 string `beschema:"1"`
	Field2 string `beschema:"2"`
}

func TestNonSequentialBeschemaTagsIssue(t *testing.T) {
	// This test reproduces the exact issue described:
	// Expected: {Sub1:{Field1:test1 Field2:test2} Sub2:{Field1:test5 Field2:2}}
	// Actual: {Sub1:{Field1:test1 Field2:test2} Sub2:{Field1:test4 Field2:1}}

	// Array data from the example: [["test1","test2","[[null]]",null,null,null,"test3"],["test4",1],["test5",2,"test6",3]]
	// Sub1 has beschema:"1" -> should map to array[0] = ["test1","test2","[[null]]",null,null,null,"test3"]
	// Sub2 has beschema:"3" -> should map to array[2] = ["test5",2,"test6",3]
	arrayData := []interface{}{
		[]interface{}{"test1", "test2", "[[null]]", nil, nil, nil, "test3"},
		[]interface{}{"test4", float64(1)},
		[]interface{}{"test5", float64(2), "test6", float64(3)},
	}

	var result Entity
	err := arrayToStruct(arrayData, &result)
	if err != nil {
		t.Fatalf("arrayToStruct failed: %v", err)
	}

	// Verify Sub1 (should get first array)
	if result.Sub1.Field1 != "test1" {
		t.Errorf("Expected Sub1.Field1 = 'test1', got '%s'", result.Sub1.Field1)
	}
	if result.Sub1.Field2 != "test2" {
		t.Errorf("Expected Sub1.Field2 = 'test2', got '%s'", result.Sub1.Field2)
	}

	// Verify Sub2 (should get third array, not second!)
	if result.Sub2.Field1 != "test5" {
		t.Errorf("Expected Sub2.Field1 = 'test5', got '%s' (this indicates the bug - getting second array instead of third)", result.Sub2.Field1)
	}
	if result.Sub2.Field2 != "2" {
		t.Errorf("Expected Sub2.Field2 = '2', got '%s' (this indicates the bug - getting second array instead of third)", result.Sub2.Field2)
	}
}

// Test structs for the new issue with modified beschema tags
type EntityModified struct {
	Sub1 SubEntity1Modified `beschema:"1"`
	Sub2 SubEntity2Modified `beschema:"3"`
}

type SubEntity1Modified struct {
	Field1 string `beschema:"1"`
	Field2 string `beschema:"2"`
}

type SubEntity2Modified struct {
	Field1 string `beschema:"3"`
	Field2 string `beschema:"4"`
}

func TestModifiedBeschemaTagsIssue(t *testing.T) {
	// This test reproduces the new issue with modified beschema tags:
	// Expected: {Sub1:{Field1:test1 Field2:test2} Sub2:{Field1:test6 Field2:3}}
	// Actual: {Sub1:{Field1:test1 Field2:test2} Sub2:{Field1:test5 Field2:2}}

	// Array data from the example: [["test1","test2","[[null]]",null,null,null,"test3"],["test4",1],["test5",2,"test6",3]]
	// Sub1 has beschema:"1" -> should map to array[0] = ["test1","test2","[[null]]",null,null,null,"test3"]
	// Sub2 has beschema:"3" -> should map to array[2] = ["test5",2,"test6",3]
	// Within Sub2's array ["test5",2,"test6",3]:
	//   - Field1 with beschema:"3" should map to index 2 -> "test6"
	//   - Field2 with beschema:"4" should map to index 3 -> 3
	arrayData := []interface{}{
		[]interface{}{"test1", "test2", "[[null]]", nil, nil, nil, "test3"},
		[]interface{}{"test4", float64(1)},
		[]interface{}{"test5", float64(2), "test6", float64(3)},
	}

	var result EntityModified
	err := arrayToStruct(arrayData, &result)
	if err != nil {
		t.Fatalf("arrayToStruct failed: %v", err)
	}

	// Verify Sub1 (should get first array)
	if result.Sub1.Field1 != "test1" {
		t.Errorf("Expected Sub1.Field1 = 'test1', got '%s'", result.Sub1.Field1)
	}
	if result.Sub1.Field2 != "test2" {
		t.Errorf("Expected Sub1.Field2 = 'test2', got '%s'", result.Sub1.Field2)
	}

	// Verify Sub2 (should get third array and map fields by beschema tags)
	if result.Sub2.Field1 != "test6" {
		t.Errorf("Expected Sub2.Field1 = 'test6', got '%s' (this indicates the bug - not using beschema tags for nested field mapping)", result.Sub2.Field1)
	}
	if result.Sub2.Field2 != "3" {
		t.Errorf("Expected Sub2.Field2 = '3', got '%s' (this indicates the bug - not using beschema tags for nested field mapping)", result.Sub2.Field2)
	}
}
