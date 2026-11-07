package repository

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
)

// jsonString returns a required string field.
func jsonString(m map[string]interface{}, key string) (string, error) {
	value, ok := m[key]
	if !ok {
		return "", fmt.Errorf("missing required field %q", key)
	}
	result, ok := value.(string)
	if !ok {
		return "", fmt.Errorf("field %q is not a string", key)
	}
	return result, nil
}

// jsonObject returns a required object field.
func jsonObject(m map[string]interface{}, key string) (map[string]interface{}, error) {
	value, ok := m[key]
	if !ok {
		return nil, fmt.Errorf("missing required field %q", key)
	}
	result, ok := value.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("field %q is not an object", key)
	}
	return result, nil
}

// jsonArray returns a required array field.
func jsonArray(m map[string]interface{}, key string) ([]interface{}, error) {
	value, ok := m[key]
	if !ok {
		return nil, fmt.Errorf("missing required field %q", key)
	}
	result, ok := value.([]interface{})
	if !ok {
		return nil, fmt.Errorf("field %q is not an array", key)
	}
	return result, nil
}

// jsonInteger returns a required integer field. Both json.Number and float64
// are accepted because callers decode with and without UseNumber.
func jsonInteger(m map[string]interface{}, key string) (int64, error) {
	value, ok := m[key]
	if !ok {
		return 0, fmt.Errorf("missing required field %q", key)
	}
	switch number := value.(type) {
	case json.Number:
		result, err := number.Int64()
		if err != nil {
			return 0, fmt.Errorf("field %q is not an integer: %w", key, err)
		}
		return result, nil
	case float64:
		if number != float64(int64(number)) {
			return 0, fmt.Errorf("field %q is not an integer", key)
		}
		return int64(number), nil
	case int:
		return int64(number), nil
	case int64:
		return number, nil
	default:
		return 0, fmt.Errorf("field %q is not a number", key)
	}
}

// jsonBoolean returns a required boolean field.
func jsonBoolean(m map[string]interface{}, key string) (bool, error) {
	value, ok := m[key]
	if !ok {
		return false, fmt.Errorf("missing required field %q", key)
	}
	result, ok := value.(bool)
	if !ok {
		return false, fmt.Errorf("field %q is not a boolean", key)
	}
	return result, nil
}

// jsonValueString reads the string held by a value object's {"value": "..."} object.
func jsonValueString(m map[string]interface{}, key string) (string, error) {
	object, err := jsonObject(m, key)
	if err != nil {
		return "", err
	}
	return jsonString(object, "value")
}

// decodeJSONObject decodes one JSON object without losing integer precision.
func decodeJSONObject(data []byte) (map[string]interface{}, error) {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	var value map[string]interface{}
	if err := decoder.Decode(&value); err != nil {
		return nil, err
	}
	if err := decoder.Decode(new(interface{})); err != io.EOF {
		if err == nil {
			err = fmt.Errorf("multiple JSON values")
		}
		return nil, err
	}
	return value, nil
}
