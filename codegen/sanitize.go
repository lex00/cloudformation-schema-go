package codegen

import (
	"unicode"
)

// SanitizeGoIdentifier ensures a string is a valid Go identifier.
//
// Rules applied:
//   - First character must be letter or underscore
//   - Subsequent characters must be letter, digit, or underscore
//   - Empty strings become "_"
//   - Go keywords get "_" suffix appended
//
// Examples:
//
//	SanitizeGoIdentifier("validName")    // "validName"
//	SanitizeGoIdentifier("123start")     // "_123start"
//	SanitizeGoIdentifier("with-dash")    // "withdash"
//	SanitizeGoIdentifier("type")         // "type_"
//	SanitizeGoIdentifier("")             // "_"
func SanitizeGoIdentifier(name string) string {
	if name == "" {
		return "_"
	}

	var result []rune
	for i, r := range name {
		if i == 0 {
			// First character must be letter or underscore
			if unicode.IsLetter(r) || r == '_' {
				result = append(result, r)
			} else if unicode.IsDigit(r) {
				// Prefix with underscore if starts with digit
				result = append(result, '_', r)
			} else {
				result = append(result, '_')
			}
		} else {
			// Subsequent characters: letter, digit, or underscore
			if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' {
				result = append(result, r)
			}
			// Skip invalid characters
		}
	}

	s := string(result)
	if s == "" {
		return "_"
	}

	// Append underscore to Go keywords
	if IsGoKeyword(s) {
		return s + "_"
	}

	return s
}
