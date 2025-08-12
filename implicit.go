package beschema

import (
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
)

type ImplicitSchema map[int]any

func MarshalImplicitSchema(schema ImplicitSchema) ([]byte, error) {
	// Convert map to ordered array based on keys
	keys := make([]int, 0, len(schema))
	for k := range schema {
		keys = append(keys, k)
	}
	sort.Ints(keys)

	// Build array in key order
	arr := make([]interface{}, len(keys))
	for i, key := range keys {
		arr[i] = schema[key]
	}

	// Marshal to JSON
	jsonData, err := json.Marshal(arr)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal to JSON: %v", err)
	}

	// Calculate size (JSON data + \r\n)
	size := len(jsonData) + 2

	// Format as "size\r\nJSON_data\r\n"
	result := fmt.Sprintf("%d\r\n%s\r\n", size, string(jsonData))

	return []byte(result), nil
}

func UnmarshalImplicitSchema(data []byte) (ImplicitSchema, error) {
	// Convert data to string
	dataStr := string(data)

	// Split by \r\n (Windows-style line breaks)
	lines := strings.Split(dataStr, "\r\n")
	if len(lines) < 2 {
		// Try splitting by \n only (Unix-style line breaks)
		lines = strings.Split(dataStr, "\n")
		if len(lines) < 2 {
			return nil, fmt.Errorf("invalid data format: expected at least 2 lines")
		}
	}

	// Parse size information from the first line
	expectedSize, err := strconv.Atoi(strings.TrimSpace(lines[0]))
	if err != nil {
		return nil, fmt.Errorf("invalid size format: %v", err)
	}

	// Parse actual JSON data from the second line
	jsonData := strings.TrimSpace(lines[1])

	// Actual data size is JSON data + \r\n (2 bytes)
	actualSize := len(jsonData) + 2
	if actualSize != expectedSize {
		return nil, fmt.Errorf("data size mismatch: expected %d, got %d", expectedSize, actualSize)
	}

	// Unmarshal to JSON array
	var arr []interface{}
	if err := json.Unmarshal([]byte(jsonData), &arr); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %v", err)
	}

	// Convert array to map with sequential keys starting from 1
	result := make(ImplicitSchema)
	for i, value := range arr {
		result[i+1] = value
	}

	return result, nil
}
