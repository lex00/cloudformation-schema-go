package codegen

// goKeywords is the set of Go language reserved keywords.
var goKeywords = map[string]bool{
	"break":       true,
	"case":        true,
	"chan":        true,
	"const":       true,
	"continue":    true,
	"default":     true,
	"defer":       true,
	"else":        true,
	"fallthrough": true,
	"for":         true,
	"func":        true,
	"go":          true,
	"goto":        true,
	"if":          true,
	"import":      true,
	"interface":   true,
	"map":         true,
	"package":     true,
	"range":       true,
	"return":      true,
	"select":      true,
	"struct":      true,
	"switch":      true,
	"type":        true,
	"var":         true,
}

// IsGoKeyword returns true if s is a Go language keyword.
func IsGoKeyword(s string) bool {
	return goKeywords[s]
}

// GoKeywords returns a copy of all Go language keywords.
func GoKeywords() []string {
	keywords := make([]string, 0, len(goKeywords))
	for k := range goKeywords {
		keywords = append(keywords, k)
	}
	return keywords
}
