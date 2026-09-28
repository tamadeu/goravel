package telego

import (
	"encoding/json"
	"strings"
)

func Redact(value any) any {
	switch typed := value.(type) {
	case map[string]any:
		redacted := make(map[string]any, len(typed))
		for key, item := range typed {
			if sensitiveKey(key) {
				redacted[key] = "[REDACTED]"
				continue
			}
			redacted[key] = Redact(item)
		}
		return redacted
	case map[string][]string:
		redacted := make(map[string]any, len(typed))
		for key, item := range typed {
			if sensitiveKey(key) {
				redacted[key] = "[REDACTED]"
				continue
			}
			redacted[key] = item
		}
		return redacted
	case map[string]string:
		redacted := make(map[string]any, len(typed))
		for key, item := range typed {
			if sensitiveKey(key) {
				redacted[key] = "[REDACTED]"
				continue
			}
			redacted[key] = item
		}
		return redacted
	case []any:
		redacted := make([]any, len(typed))
		for index, item := range typed {
			redacted[index] = Redact(item)
		}
		return redacted
	default:
		return value
	}
}

func JSON(value any, limit int) string {
	data, err := json.Marshal(Redact(value))
	if err != nil {
		return ""
	}
	if limit > 0 && len(data) > limit {
		return string(data[:limit]) + "..."
	}

	return string(data)
}

func sensitiveKey(key string) bool {
	lowerKey := strings.ToLower(key)
	for _, sensitive := range []string{"authorization", "cookie", "password", "passwd", "secret", "token", "api_key", "apikey"} {
		if strings.Contains(lowerKey, sensitive) {
			return true
		}
	}

	return false
}
