package beschema

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

type ImplicitSchema []any

func MarshalImplicitSchema(schema ImplicitSchema) ([]byte, error) {
	// Marshal slice directly to JSON
	jsonData, err := json.Marshal(schema)
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

	// Unmarshal to JSON array and return as ImplicitSchema
	var result ImplicitSchema
	if err := json.Unmarshal([]byte(jsonData), &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %v", err)
	}

	return result, nil
}
