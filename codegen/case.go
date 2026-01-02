package codegen

import (
	"strings"
	"unicode"
)

// ToSnakeCase converts PascalCase or camelCase to snake_case.
// e.g., "BucketName" -> "bucket_name", "getHTTPResponse" -> "get_h_t_t_p_response"
func ToSnakeCase(s string) string {
	var result strings.Builder
	for i, r := range s {
		if i > 0 && unicode.IsUpper(r) {
			result.WriteRune('_')
		}
		result.WriteRune(unicode.ToLower(r))
	}
	return result.String()
}

// ToPascalCase converts snake_case, kebab-case, or space-separated to PascalCase.
// e.g., "bucket_name" -> "BucketName", "my-function" -> "MyFunction"
func ToPascalCase(s string) string {
	var result strings.Builder
	capitalizeNext := true

	for _, r := range s {
		if r == '_' || r == '-' || r == ' ' {
			capitalizeNext = true
			continue
		}
		if capitalizeNext {
			result.WriteRune(unicode.ToUpper(r))
			capitalizeNext = false
		} else {
			result.WriteRune(unicode.ToLower(r))
		}
	}

	return result.String()
}
