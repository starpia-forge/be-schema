package beschema

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

func MarshalExplicitSchema[T any](v T) ([]byte, error) {
	// 구조체를 배열로 변환
	arr, err := structToArray(v)
	if err != nil {
		return nil, err
	}

	// JSON으로 마셜링
	jsonData, err := json.Marshal(arr)
	if err != nil {
		return nil, err
	}

	// 크기 정보와 함께 결합 (JSON 데이터 + \r\n 2바이트)
	size := len(jsonData) + 2
	result := fmt.Sprintf("%d\r\n%s\r\n", size, string(jsonData))

	return []byte(result), nil
}

func UnmarshalExplicitSchema[T any](data []byte) (T, error) {
	var result T

	// 전체 데이터를 문자열로 변환
	dataStr := string(data)

	// \r\n으로 분할 (Windows 스타일 줄바꿈)
	lines := strings.Split(dataStr, "\r\n")
	if len(lines) < 2 {
		// \n으로만 분할 시도 (Unix 스타일 줄바꿈)
		lines = strings.Split(dataStr, "\n")
		if len(lines) < 2 {
			return result, fmt.Errorf("invalid data format: expected at least 2 lines")
		}
	}

	// 첫 번째 줄에서 크기 정보 파싱
	expectedSize, err := strconv.Atoi(strings.TrimSpace(lines[0]))
	if err != nil {
		return result, fmt.Errorf("invalid size format: %v", err)
	}

	// 두 번째 줄에서 실제 JSON 데이터 파싱
	jsonData := strings.TrimSpace(lines[1])

	// 실제 데이터 크기는 JSON 데이터 + \r\n (2바이트)
	actualSize := len(jsonData) + 2
	if actualSize != expectedSize {
		return result, fmt.Errorf("data size mismatch: expected %d, got %d (JSON: %d + CRLF: 2)", expectedSize, actualSize, len(jsonData))
	}

	// JSON 배열로 언마셜링
	var arr []interface{}
	if err := json.Unmarshal([]byte(jsonData), &arr); err != nil {
		return result, fmt.Errorf("failed to unmarshal JSON: %v", err)
	}

	// 배열을 구조체로 변환
	if err := arrayToStruct(arr, &result); err != nil {
		return result, err
	}

	return result, nil
}

// 구조체를 배열로 변환하는 헬퍼 함수
func structToArray(v interface{}) ([]interface{}, error) {
	val := reflect.ValueOf(v)
	typ := reflect.TypeOf(v)

	// 포인터인 경우 역참조
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

		// unexported 필드는 건너뛰기
		if !field.CanInterface() {
			continue
		}

		// 필드가 구조체인 경우 재귀적으로 처리
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

// 배열을 구조체로 변환하는 헬퍼 함수
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

	// 구조체의 필드들을 순회하면서 배열 데이터와 매핑
	arrayIndex := 0

	for i := 0; i < val.NumField() && arrayIndex < len(arr); i++ {
		field := val.Field(i)
		fieldType := typ.Field(i)

		// unexported 필드는 건너뛰기
		if !field.CanSet() {
			continue
		}

		if arrayIndex >= len(arr) {
			break
		}

		// 현재 배열 요소 가져오기
		arrValue := arr[arrayIndex]

		// 필드가 구조체인 경우
		if field.Kind() == reflect.Struct {
			// 배열 데이터가 슬라이스인지 확인
			if subArr, ok := arrValue.([]interface{}); ok {
				// 구조체의 각 필드를 배열 요소와 매핑
				if err := populateStructFromArray(field, subArr); err != nil {
					return fmt.Errorf("failed to populate struct field %s: %v", fieldType.Name, err)
				}
			} else {
				return fmt.Errorf("expected array for struct field %s, got %T", fieldType.Name, arrValue)
			}
		} else {
			// 기본 타입 필드 설정
			if err := setFieldValue(field, arrValue); err != nil {
				return fmt.Errorf("failed to set field %s: %v", fieldType.Name, err)
			}
		}

		arrayIndex++
	}

	return nil
}

// 배열에서 구조체 필드들을 채우는 헬퍼 함수
func populateStructFromArray(structVal reflect.Value, arr []interface{}) error {
	structType := structVal.Type()

	for i := 0; i < structVal.NumField() && i < len(arr); i++ {
		field := structVal.Field(i)

		if !field.CanSet() {
			continue
		}

		arrValue := arr[i]

		if field.Kind() == reflect.Struct {
			// 중첩 구조체인 경우
			if subArr, ok := arrValue.([]interface{}); ok {
				if err := populateStructFromArray(field, subArr); err != nil {
					return fmt.Errorf("failed to populate nested struct field %s: %v", structType.Field(i).Name, err)
				}
			}
		} else {
			// 기본 타입 필드 설정
			if err := setFieldValue(field, arrValue); err != nil {
				return fmt.Errorf("failed to set field %s: %v", structType.Field(i).Name, err)
			}
		}
	}

	return nil
}

// 필드 값을 설정하는 헬퍼 함수
func setFieldValue(field reflect.Value, value interface{}) error {
	if value == nil {
		return nil // nil 값은 무시
	}

	fieldType := field.Type()
	valueType := reflect.TypeOf(value)

	// 타입이 직접 일치하는 경우
	if valueType == fieldType {
		field.Set(reflect.ValueOf(value))
		return nil
	}

	// 타입 변환이 필요한 경우
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
