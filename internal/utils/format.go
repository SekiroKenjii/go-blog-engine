package utils

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/go-playground/validator/v10"
)

// FormatFieldError replaces placeholders in a string with values from a validator.FieldError
func FormatFieldError(format string, f validator.FieldError) string {
	replacements := buildPlaceholderMap(format, f)

	for placeholder, value := range replacements {
		format = strings.ReplaceAll(format, placeholder, value)
	}

	return format
}

// extractPlaceholders finds all placeholders in the input string
func extractPlaceholders(input string) []string {
	re := regexp.MustCompile(`__[^_]+__`)

	return re.FindAllString(input, -1)
}

// buildPlaceholderMap creates a map of placeholders and their corresponding values
func buildPlaceholderMap(msg string, fe validator.FieldError) map[string]string {
	placeholders := extractPlaceholders(msg)

	result := make(map[string]string)

	for _, ph := range placeholders {
		switch ph {
		case "__FIELD__":
			result[ph] = fe.Field()
		case "__PARAM__":
			result[ph] = fe.Param()
		case "__TAG__":
			result[ph] = fe.Tag()
		case "__VALUE__":
			result[ph] = toString(fe.Value())
		case "__TYPE__":
			result[ph] = toString(fe.Type())
		case "__STRUCT_FIELD__":
			result[ph] = fe.StructField()
		case "__STRUCT_NAME__":
			result[ph] = fe.StructNamespace()
		default:
			result[ph] = ""
		}
	}

	return result
}

// toString safely converts any value to a string
func toString(value any) string {
	if value == nil {
		return ""
	}

	return strings.TrimSpace(strings.ToLower(fmt.Sprintf("%v", value)))
}
