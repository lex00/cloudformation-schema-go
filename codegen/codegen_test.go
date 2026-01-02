package codegen

import (
	"reflect"
	"sort"
	"testing"
)

func TestToSnakeCase(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"BucketName", "bucket_name"},
		{"S3Bucket", "s3_bucket"},
		{"HTTPResponse", "h_t_t_p_response"},
		{"getAZs", "get_a_zs"},
		{"simple", "simple"},
		{"", ""},
		{"A", "a"},
		{"AB", "a_b"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := ToSnakeCase(tt.input)
			if result != tt.expected {
				t.Errorf("ToSnakeCase(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestToPascalCase(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"bucket_name", "BucketName"},
		{"my-function", "MyFunction"},
		{"hello world", "HelloWorld"},
		{"simple", "Simple"},
		{"ALLCAPS", "Allcaps"},
		{"", ""},
		{"a", "A"},
		{"a_b_c", "ABC"},
		{"__leading", "Leading"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := ToPascalCase(tt.input)
			if result != tt.expected {
				t.Errorf("ToPascalCase(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestIsGoKeyword(t *testing.T) {
	keywords := []string{
		"break", "case", "chan", "const", "continue",
		"default", "defer", "else", "fallthrough", "for",
		"func", "go", "goto", "if", "import",
		"interface", "map", "package", "range", "return",
		"select", "struct", "switch", "type", "var",
	}

	for _, kw := range keywords {
		if !IsGoKeyword(kw) {
			t.Errorf("IsGoKeyword(%q) = false, want true", kw)
		}
	}

	nonKeywords := []string{
		"foo", "bar", "Bucket", "main", "init", "string", "int",
	}

	for _, nk := range nonKeywords {
		if IsGoKeyword(nk) {
			t.Errorf("IsGoKeyword(%q) = true, want false", nk)
		}
	}
}

func TestGoKeywords(t *testing.T) {
	keywords := GoKeywords()
	if len(keywords) != 25 {
		t.Errorf("GoKeywords() returned %d keywords, want 25", len(keywords))
	}

	// Check that all returned keywords are valid
	for _, kw := range keywords {
		if !IsGoKeyword(kw) {
			t.Errorf("GoKeywords() returned %q which is not a keyword", kw)
		}
	}
}

func TestSanitizeGoIdentifier(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"validName", "validName"},
		{"Valid123", "Valid123"},
		{"_underscore", "_underscore"},
		{"123start", "_123start"},
		{"with-dash", "withdash"},
		{"with.dot", "withdot"},
		{"with space", "withspace"},
		{"type", "type_"},
		{"var", "var_"},
		{"func", "func_"},
		{"", "_"},
		{"!!!", "_"},
		{"@#$%", "_"},
		{"a@b", "ab"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := SanitizeGoIdentifier(tt.input)
			if result != tt.expected {
				t.Errorf("SanitizeGoIdentifier(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestTopologicalSort(t *testing.T) {
	t.Run("simple chain", func(t *testing.T) {
		// A -> B -> C (A depends on B, B depends on C)
		nodes := []string{"A", "B", "C"}
		deps := map[string][]string{
			"A": {"B"},
			"B": {"C"},
			"C": {},
		}
		getDeps := func(n string) []string { return deps[n] }

		result := TopologicalSort(nodes, getDeps)
		expected := []string{"C", "B", "A"}

		if !reflect.DeepEqual(result, expected) {
			t.Errorf("TopologicalSort() = %v, want %v", result, expected)
		}
	})

	t.Run("multiple roots", func(t *testing.T) {
		// A -> C, B -> C
		nodes := []string{"A", "B", "C"}
		deps := map[string][]string{
			"A": {"C"},
			"B": {"C"},
			"C": {},
		}
		getDeps := func(n string) []string { return deps[n] }

		result := TopologicalSort(nodes, getDeps)

		// C must come first
		if result[0] != "C" {
			t.Errorf("TopologicalSort() first element = %v, want C", result[0])
		}
		// A and B should follow (in sorted order)
		remaining := result[1:]
		sort.Strings(remaining)
		if !reflect.DeepEqual(remaining, []string{"A", "B"}) {
			t.Errorf("TopologicalSort() remaining = %v, want [A B]", remaining)
		}
	})

	t.Run("no dependencies", func(t *testing.T) {
		nodes := []string{"C", "A", "B"}
		deps := map[string][]string{
			"A": {},
			"B": {},
			"C": {},
		}
		getDeps := func(n string) []string { return deps[n] }

		result := TopologicalSort(nodes, getDeps)
		expected := []string{"A", "B", "C"} // Sorted alphabetically

		if !reflect.DeepEqual(result, expected) {
			t.Errorf("TopologicalSort() = %v, want %v", result, expected)
		}
	})

	t.Run("cycle handling", func(t *testing.T) {
		// A -> B -> C -> A (cycle)
		nodes := []string{"A", "B", "C"}
		deps := map[string][]string{
			"A": {"B"},
			"B": {"C"},
			"C": {"A"},
		}
		getDeps := func(n string) []string { return deps[n] }

		result := TopologicalSort(nodes, getDeps)

		// All nodes should be in result
		if len(result) != 3 {
			t.Errorf("TopologicalSort() len = %d, want 3", len(result))
		}

		// Check all nodes are present
		resultSet := make(map[string]bool)
		for _, n := range result {
			resultSet[n] = true
		}
		for _, n := range nodes {
			if !resultSet[n] {
				t.Errorf("TopologicalSort() missing node %s", n)
			}
		}
	})

	t.Run("empty input", func(t *testing.T) {
		result := TopologicalSort(nil, func(string) []string { return nil })
		if len(result) != 0 {
			t.Errorf("TopologicalSort(nil) = %v, want []", result)
		}
	})
}
