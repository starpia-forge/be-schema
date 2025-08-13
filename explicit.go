package beschema

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
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

// fieldInfo holds information about a struct field and its beschema tag
type fieldInfo struct {
	field     reflect.Value
	fieldType reflect.StructField
	tagValue  int
}

// structToArray is a helper function that converts a struct to an array representation.
// It recursively processes nested structs and handles unexported fields appropriately.
// Fields are ordered by their beschema tag values.
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

	// Collect field information with beschema tags
	var fields []fieldInfo
	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		fieldType := typ.Field(i)

		// Skip unexported fields
		if !field.CanInterface() {
			continue
		}

		// Parse beschema tag
		tagValue := i + 1 // default to field order (1-based)
		if tag := fieldType.Tag.Get("beschema"); tag != "" {
			if parsedTag, err := strconv.Atoi(tag); err == nil {
				tagValue = parsedTag
			}
		}

		fields = append(fields, fieldInfo{
			field:     field,
			fieldType: fieldType,
			tagValue:  tagValue,
		})
	}

	// Sort fields by beschema tag value
	sort.Slice(fields, func(i, j int) bool {
		return fields[i].tagValue < fields[j].tagValue
	})

	// Find the maximum tag value to determine array size
	maxTagValue := 0
	for _, fieldInfo := range fields {
		if fieldInfo.tagValue > maxTagValue {
			maxTagValue = fieldInfo.tagValue
		}
	}

	// Create result array with proper size, initialized with nulls
	result := make([]interface{}, maxTagValue)
	for i := range result {
		result[i] = nil
	}

	// Place each field at its correct index (tagValue - 1)
	for _, fieldInfo := range fields {
		arrayIndex := fieldInfo.tagValue - 1 // Convert 1-based tag to 0-based array index
		if arrayIndex < 0 || arrayIndex >= len(result) {
			continue // Skip if tag value is out of bounds
		}

		// If a field is a struct, process recursively
		if fieldInfo.field.Kind() == reflect.Struct {
			subArray, err := structToArray(fieldInfo.field.Interface())
			if err != nil {
				return nil, fmt.Errorf("failed to convert field %s: %v", fieldInfo.fieldType.Name, err)
			}
			result[arrayIndex] = subArray
		} else {
			result[arrayIndex] = fieldInfo.field.Interface()
		}
	}

	return result, nil
}

// arrayToStruct is a helper function that converts an array to a struct.
// The target parameter must be a pointer to the struct to be populated.
// Fields are mapped based on their beschema tag values.
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

	// Collect field information with beschema tags
	var fields []fieldInfo
	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		fieldType := typ.Field(i)

		// Skip unexported fields
		if !field.CanSet() {
			continue
		}

		// Parse beschema tag
		tagValue := i + 1 // default to field order (1-based)
		if tag := fieldType.Tag.Get("beschema"); tag != "" {
			if parsedTag, err := strconv.Atoi(tag); err == nil {
				tagValue = parsedTag
			}
		}

		fields = append(fields, fieldInfo{
			field:     field,
			fieldType: fieldType,
			tagValue:  tagValue,
		})
	}

	// Sort fields by beschema tag value
	sort.Slice(fields, func(i, j int) bool {
		return fields[i].tagValue < fields[j].tagValue
	})

	// Map array elements to fields based on tag values (1-based to 0-based conversion)
	for _, fieldInfo := range fields {
		arrayIndex := fieldInfo.tagValue - 1 // Convert 1-based tag to 0-based array index
		if arrayIndex < 0 || arrayIndex >= len(arr) {
			continue // Skip if tag value is out of bounds
		}

		arrValue := arr[arrayIndex]

		// If the field is a struct
		if fieldInfo.field.Kind() == reflect.Struct {
			// Check if array data is a slice
			if subArr, ok := arrValue.([]interface{}); ok {
				// Map each field of the struct with array elements
				if err := populateStructFromArray(fieldInfo.field, subArr); err != nil {
					return fmt.Errorf("failed to populate struct field %s: %v", fieldInfo.fieldType.Name, err)
				}
			} else {
				return fmt.Errorf("expected array for struct field %s, got %T", fieldInfo.fieldType.Name, arrValue)
			}
		} else {
			// Set a basic type field
			if err := setFieldValue(fieldInfo.field, arrValue); err != nil {
				return fmt.Errorf("failed to set field %s: %v", fieldInfo.fieldType.Name, err)
			}
		}
	}

	return nil
}

// populateStructFromArray is a helper function that populates struct fields from an array.
// It handles nested structs recursively and converts array elements to appropriate field types.
// Fields are mapped based on their beschema tag values.
func populateStructFromArray(structVal reflect.Value, arr []interface{}) error {
	structType := structVal.Type()

	// Collect field information with beschema tags
	var fields []fieldInfo
	for i := 0; i < structVal.NumField(); i++ {
		field := structVal.Field(i)
		fieldType := structType.Field(i)

		if !field.CanSet() {
			continue
		}

		// Parse beschema tag
		tagValue := i + 1 // default to field order (1-based)
		if tag := fieldType.Tag.Get("beschema"); tag != "" {
			if parsedTag, err := strconv.Atoi(tag); err == nil {
				tagValue = parsedTag
			}
		}

		fields = append(fields, fieldInfo{
			field:     field,
			fieldType: fieldType,
			tagValue:  tagValue,
		})
	}

	// Sort fields by beschema tag value
	sort.Slice(fields, func(i, j int) bool {
		return fields[i].tagValue < fields[j].tagValue
	})

	// Map array elements to fields based on tag values (1-based to 0-based conversion)
	for _, fieldInfo := range fields {
		arrayIndex := fieldInfo.tagValue - 1 // Convert 1-based tag to 0-based array index
		if arrayIndex < 0 || arrayIndex >= len(arr) {
			continue // Skip if tag value is out of bounds
		}

		arrValue := arr[arrayIndex]

		if fieldInfo.field.Kind() == reflect.Struct {
			// For nested structs
			if subArr, ok := arrValue.([]interface{}); ok {
				if err := populateStructFromArray(fieldInfo.field, subArr); err != nil {
					return fmt.Errorf("failed to populate nested struct field %s: %v", fieldInfo.fieldType.Name, err)
				}
			}
		} else {
			// Set a basic type field
			if err := setFieldValue(fieldInfo.field, arrValue); err != nil {
				return fmt.Errorf("failed to set field %s: %v", fieldInfo.fieldType.Name, err)
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
