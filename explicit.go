package beschema

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// MarshalExplicitSchema converts a struct to a byte array following the explicit schema format.
// It converts the struct to an array representation, marshals it to JSON,
// and prepends size information in the format: "size\r\nJSON_data\r\n".
func MarshalExplicitSchema[T any](v T) ([]byte, error) {
	// Convert struct to array
	arr, err := structToArray(v)
	if err != nil {
		return nil, err
	}

	// Marshal to JSON
	jsonData, err := json.Marshal(arr)
	if err != nil {
		return nil, err
	}

	// Combine with size information (JSON data + \r\n 2 bytes)
	size := len(jsonData) + 2
	result := fmt.Sprintf("%d\r\n%s\r\n", size, string(jsonData))

	return []byte(result), nil
}

// UnmarshalExplicitSchema parses byte data in an explicit schema format and converts it to the specified struct type.
// The input data should be in the format: "size\r\nJSON_data\r\n".
// It validates the size information and converts the JSON array back to the target struct.
func UnmarshalExplicitSchema[T any](data []byte) (T, error) {
	var result T

	// Convert entire data to string
	dataStr := string(data)

	// Split by \r\n (Windows-style line breaks)
	lines := strings.Split(dataStr, "\r\n")
	if len(lines) < 2 {
		// Try splitting by \n only (Unix-style line breaks)
		lines = strings.Split(dataStr, "\n")
		if len(lines) < 2 {
			return result, fmt.Errorf("invalid data format: expected at least 2 lines")
		}
	}

	// Parse size information from the first line
	expectedSize, err := strconv.Atoi(strings.TrimSpace(lines[0]))
	if err != nil {
		return result, fmt.Errorf("invalid size format: %v", err)
	}

	// Parse actual JSON data from the second line
	jsonData := strings.TrimSpace(lines[1])

	// Actual data size is JSON data + \r\n (2 bytes)
	actualSize := len(jsonData) + 2
	if actualSize != expectedSize {
		return result, fmt.Errorf("data size mismatch: expected %d, got %d (JSON: %d + CRLF: 2)", expectedSize, actualSize, len(jsonData))
	}

	// Unmarshal to JSON array
	var arr []interface{}
	if err := json.Unmarshal([]byte(jsonData), &arr); err != nil {
		return result, fmt.Errorf("failed to unmarshal JSON: %v", err)
	}

	// Convert array to struct
	if err := arrayToStruct(arr, &result); err != nil {
		return result, err
	}

	return result, nil
}

// structToArray is a helper function that converts a struct to an array representation.
// It recursively processes nested structs and handles unexported fields appropriately.
func structToArray(v interface{}) ([]interface{}, error) {
	val := reflect.ValueOf(v)
	typ := reflect.TypeOf(v)

	// Dereference if it's a pointer
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
		typ = typ.Elem()
	}

	if val.Kind() != reflect.Struct {
		return nil, fmt.Errorf("expected struct, got %s", val.Kind())
	}

	var result []interface{}

	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		fieldType := typ.Field(i)

		// Skip unexported fields
		if !field.CanInterface() {
			continue
		}

		// If a field is a struct, process recursively
		if field.Kind() == reflect.Struct {
			subArray, err := structToArray(field.Interface())
			if err != nil {
				return nil, fmt.Errorf("failed to convert field %s: %v", fieldType.Name, err)
			}
			result = append(result, subArray)
		} else {
			result = append(result, field.Interface())
		}
	}

	return result, nil
}

// arrayToStruct is a helper function that converts an array to a struct.
// The target parameter must be a pointer to the struct to be populated.
func arrayToStruct(arr []interface{}, target interface{}) error {
	val := reflect.ValueOf(target)
	if val.Kind() != reflect.Ptr {
		return fmt.Errorf("target must be a pointer")
	}

	val = val.Elem()
	typ := val.Type()

	if val.Kind() != reflect.Struct {
		return fmt.Errorf("target must be a pointer to struct")
	}

	// Iterate through struct fields and map them with array data
	arrayIndex := 0

	for i := 0; i < val.NumField() && arrayIndex < len(arr); i++ {
		field := val.Field(i)
		fieldType := typ.Field(i)

		// Skip unexported fields
		if !field.CanSet() {
			continue
		}

		if arrayIndex >= len(arr) {
			break
		}

		// Get current array element
		arrValue := arr[arrayIndex]

		// If the field is a struct
		if field.Kind() == reflect.Struct {
			// Check if array data is a slice
			if subArr, ok := arrValue.([]interface{}); ok {
				// Map each field of the struct with array elements
				if err := populateStructFromArray(field, subArr); err != nil {
					return fmt.Errorf("failed to populate struct field %s: %v", fieldType.Name, err)
				}
			} else {
				return fmt.Errorf("expected array for struct field %s, got %T", fieldType.Name, arrValue)
			}
		} else {
			// Set a basic type field
			if err := setFieldValue(field, arrValue); err != nil {
				return fmt.Errorf("failed to set field %s: %v", fieldType.Name, err)
			}
		}

		arrayIndex++
	}

	return nil
}

// populateStructFromArray is a helper function that populates struct fields from an array.
// It handles nested structs recursively and converts array elements to appropriate field types.
func populateStructFromArray(structVal reflect.Value, arr []interface{}) error {
	structType := structVal.Type()

	for i := 0; i < structVal.NumField() && i < len(arr); i++ {
		field := structVal.Field(i)

		if !field.CanSet() {
			continue
		}

		arrValue := arr[i]

		if field.Kind() == reflect.Struct {
			// For nested structs
			if subArr, ok := arrValue.([]interface{}); ok {
				if err := populateStructFromArray(field, subArr); err != nil {
					return fmt.Errorf("failed to populate nested struct field %s: %v", structType.Field(i).Name, err)
				}
			}
		} else {
			// Set a basic type field
			if err := setFieldValue(field, arrValue); err != nil {
				return fmt.Errorf("failed to set field %s: %v", structType.Field(i).Name, err)
			}
		}
	}

	return nil
}

// setFieldValue is a helper function that sets a field value with an appropriate type conversion.
// It handles type conversions between interface{} values and struct field types,
// supporting string, numeric, and boolean types.
func setFieldValue(field reflect.Value, value interface{}) error {
	if value == nil {
		return nil // Ignore nil values
	}

	fieldType := field.Type()
	valueType := reflect.TypeOf(value)

	// If types match directly
	if valueType == fieldType {
		field.Set(reflect.ValueOf(value))
		return nil
	}

	// Type conversion is needed
	switch fieldType.Kind() {
	case reflect.String:
		if str, ok := value.(string); ok {
			field.SetString(str)
		} else {
			field.SetString(fmt.Sprintf("%v", value))
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if num, ok := value.(float64); ok {
			field.SetInt(int64(num))
		} else if str, ok := value.(string); ok {
			if intVal, err := strconv.ParseInt(str, 10, 64); err == nil {
				field.SetInt(intVal)
			}
		}
	case reflect.Float32, reflect.Float64:
		if num, ok := value.(float64); ok {
			field.SetFloat(num)
		} else if str, ok := value.(string); ok {
			if floatVal, err := strconv.ParseFloat(str, 64); err == nil {
				field.SetFloat(floatVal)
			}
		}
	case reflect.Bool:
		if b, ok := value.(bool); ok {
			field.SetBool(b)
		} else if str, ok := value.(string); ok {
			if boolVal, err := strconv.ParseBool(str); err == nil {
				field.SetBool(boolVal)
			}
		}
	default:
		return fmt.Errorf("unsupported field type: %s", fieldType.Kind())
	}

	return nil
}
