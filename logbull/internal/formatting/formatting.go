package formatting

import (
	"encoding/json"
	"fmt"
	"strings"
)

const defaultMaxMessageLength = 10_000

func FormatMessage(message string) string {
	message = strings.TrimSpace(message)
	if len(message) > defaultMaxMessageLength {
		message = message[:defaultMaxMessageLength-3] + "..."
	}
	return message
}

func EnsureFields(fields map[string]any) map[string]any {
	if fields == nil {
		return make(map[string]any)
	}

	formatted := make(map[string]any)
	for key, value := range fields {
		key = strings.TrimSpace(key)
		if key == "" {
			continue
		}

		if isJSONSerializable(value) {
			formatted[key] = value
		} else {
			formatted[key] = convertToString(value)
		}
	}

	return formatted
}

func MergeFields(base, additional map[string]any) map[string]any {
	result := EnsureFields(base)
	for key, value := range EnsureFields(additional) {
		result[key] = value
	}
	return result
}

func isJSONSerializable(value any) bool {
	_, err := json.Marshal(value)
	return err == nil
}

func convertToString(value any) string {
	if value == nil {
		return "null"
	}

	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Sprint(value)
	}
	return string(data)
}
