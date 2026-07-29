package validator

import (
	"errors"
	"fmt"
	"strings"
	"unicode"

	playgroundValidator "github.com/go-playground/validator/v10"
)

// ValidationMessages converts Gin validation errors into messages keyed by the
// JSON-style field name. The bool result distinguishes validation errors from
// malformed JSON and other binding errors.
func ValidationMessages(err error) (map[string]string, bool) {
	var validationErrors playgroundValidator.ValidationErrors
	if !errors.As(err, &validationErrors) {
		return nil, false
	}

	messages := make(map[string]string, len(validationErrors))
	for _, fieldError := range validationErrors {
		field := toSnakeCase(fieldError.Field())

		switch fieldError.Tag() {
		case "required":
			messages[field] = fmt.Sprintf("%s is required", field)
		default:
			messages[field] = fmt.Sprintf("%s is invalid", field)
		}
	}

	return messages, true
}

func toSnakeCase(value string) string {
	var result strings.Builder
	runes := []rune(value)

	for i, current := range runes {
		if unicode.IsUpper(current) {
			if i > 0 && (unicode.IsLower(runes[i-1]) ||
				(i+1 < len(runes) && unicode.IsLower(runes[i+1]))) {
				result.WriteByte('_')
			}
			result.WriteRune(unicode.ToLower(current))
			continue
		}
		result.WriteRune(current)
	}

	return result.String()
}
