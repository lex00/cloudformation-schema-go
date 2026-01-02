// Package codegen provides utilities for code generation.
//
// This package extracts common code generation utilities used across
// wetwire-aws and cfn-lint-go for consistent identifier handling,
// case conversion, and topological sorting.
//
// # Case Conversion
//
// Convert between naming conventions:
//
//	snake := ToSnakeCase("BucketName")  // "bucket_name"
//	pascal := ToPascalCase("bucket_name") // "BucketName"
//
// # Go Keywords
//
// Check and handle Go reserved keywords:
//
//	if IsGoKeyword("type") { ... }
//
// # Identifier Sanitization
//
// Ensure strings are valid Go identifiers:
//
//	name := SanitizeGoIdentifier("123-invalid") // "_123invalid"
//
// # Topological Sorting
//
// Sort nodes by dependencies using Kahn's algorithm:
//
//	sorted := TopologicalSort(nodes, getDeps)
package codegen
