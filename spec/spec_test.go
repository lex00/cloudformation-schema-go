package spec_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/lex00/cloudformation-schema-go/spec"
)

// TestSpecJSON is a minimal valid CloudFormation spec for testing.
const testSpecJSON = `{
	"ResourceSpecificationVersion": "1.0.0",
	"ResourceTypes": {
		"AWS::S3::Bucket": {
			"Documentation": "S3 bucket resource",
			"Attributes": {
				"Arn": {
					"PrimitiveType": "String"
				},
				"DomainName": {
					"PrimitiveType": "String"
				}
			},
			"Properties": {
				"BucketName": {
					"Documentation": "Name of the bucket",
					"Required": false,
					"PrimitiveType": "String",
					"UpdateType": "Immutable"
				},
				"Tags": {
					"Documentation": "Resource tags",
					"Required": false,
					"Type": "List",
					"ItemType": "Tag",
					"UpdateType": "Mutable"
				},
				"CorsConfiguration": {
					"Documentation": "CORS configuration",
					"Required": false,
					"Type": "CorsConfiguration",
					"UpdateType": "Mutable"
				}
			}
		},
		"AWS::EC2::Instance": {
			"Documentation": "EC2 instance resource",
			"Properties": {
				"InstanceType": {
					"Required": true,
					"PrimitiveType": "String"
				},
				"ImageId": {
					"Required": true,
					"PrimitiveType": "String"
				}
			}
		}
	},
	"PropertyTypes": {
		"AWS::S3::Bucket.CorsConfiguration": {
			"Documentation": "CORS configuration for bucket",
			"Properties": {
				"CorsRules": {
					"Required": true,
					"Type": "List",
					"ItemType": "CorsRule"
				}
			}
		},
		"AWS::S3::Bucket.Tag": {
			"Documentation": "Resource tag",
			"Properties": {
				"Key": {
					"Required": true,
					"PrimitiveType": "String"
				},
				"Value": {
					"Required": true,
					"PrimitiveType": "String"
				}
			}
		}
	}
}`

func loadTestSpec(t *testing.T) *spec.Spec {
	t.Helper()
	var s spec.Spec
	if err := json.Unmarshal([]byte(testSpecJSON), &s); err != nil {
		t.Fatalf("failed to unmarshal test spec: %v", err)
	}
	return &s
}

func TestSpec_GetResourceType(t *testing.T) {
	s := loadTestSpec(t)

	tests := []struct {
		name     string
		typeName string
		want     bool
	}{
		{"exists", "AWS::S3::Bucket", true},
		{"exists_ec2", "AWS::EC2::Instance", true},
		{"not_exists", "AWS::S3::Object", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rt := s.GetResourceType(tt.typeName)
			if (rt != nil) != tt.want {
				t.Errorf("GetResourceType(%q) = %v, want %v", tt.typeName, rt != nil, tt.want)
			}
		})
	}
}

func TestSpec_GetPropertyType(t *testing.T) {
	s := loadTestSpec(t)

	tests := []struct {
		name     string
		typeName string
		want     bool
	}{
		{"exists", "AWS::S3::Bucket.CorsConfiguration", true},
		{"exists_tag", "AWS::S3::Bucket.Tag", true},
		{"not_exists", "AWS::S3::Bucket.NotExists", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pt := s.GetPropertyType(tt.typeName)
			if (pt != nil) != tt.want {
				t.Errorf("GetPropertyType(%q) = %v, want %v", tt.typeName, pt != nil, tt.want)
			}
		})
	}
}

func TestSpec_HasResourceType(t *testing.T) {
	s := loadTestSpec(t)

	if !s.HasResourceType("AWS::S3::Bucket") {
		t.Error("expected HasResourceType to return true for AWS::S3::Bucket")
	}
	if s.HasResourceType("AWS::NotReal::Resource") {
		t.Error("expected HasResourceType to return false for non-existent resource")
	}
}

func TestSpec_ResourceTypeNames(t *testing.T) {
	s := loadTestSpec(t)

	names := s.ResourceTypeNames()
	if len(names) != 2 {
		t.Errorf("expected 2 resource types, got %d", len(names))
	}

	// Check that both expected types are present
	found := make(map[string]bool)
	for _, name := range names {
		found[name] = true
	}
	if !found["AWS::S3::Bucket"] || !found["AWS::EC2::Instance"] {
		t.Errorf("missing expected resource types in %v", names)
	}
}

func TestResourceType_GetRequiredProperties(t *testing.T) {
	s := loadTestSpec(t)

	// S3 Bucket has no required properties
	bucket := s.GetResourceType("AWS::S3::Bucket")
	if bucket == nil {
		t.Fatal("expected bucket resource type")
	}
	required := bucket.GetRequiredProperties()
	if len(required) != 0 {
		t.Errorf("expected 0 required properties for bucket, got %d", len(required))
	}

	// EC2 Instance has required properties
	instance := s.GetResourceType("AWS::EC2::Instance")
	if instance == nil {
		t.Fatal("expected instance resource type")
	}
	required = instance.GetRequiredProperties()
	if len(required) != 2 {
		t.Errorf("expected 2 required properties for instance, got %d", len(required))
	}
}

func TestResourceType_GetProperty(t *testing.T) {
	s := loadTestSpec(t)
	bucket := s.GetResourceType("AWS::S3::Bucket")

	prop := bucket.GetProperty("BucketName")
	if prop == nil {
		t.Fatal("expected BucketName property")
	}
	if prop.PrimitiveType != "String" {
		t.Errorf("expected PrimitiveType String, got %s", prop.PrimitiveType)
	}

	prop = bucket.GetProperty("NotExists")
	if prop != nil {
		t.Error("expected nil for non-existent property")
	}
}

func TestResourceType_GetAttribute(t *testing.T) {
	s := loadTestSpec(t)
	bucket := s.GetResourceType("AWS::S3::Bucket")

	attr := bucket.GetAttribute("Arn")
	if attr == nil {
		t.Fatal("expected Arn attribute")
	}
	if attr.PrimitiveType != "String" {
		t.Errorf("expected PrimitiveType String, got %s", attr.PrimitiveType)
	}

	if !bucket.HasAttribute("Arn") {
		t.Error("expected HasAttribute to return true for Arn")
	}
	if bucket.HasAttribute("NotExists") {
		t.Error("expected HasAttribute to return false for non-existent attribute")
	}
}

func TestProperty_TypeChecks(t *testing.T) {
	s := loadTestSpec(t)
	bucket := s.GetResourceType("AWS::S3::Bucket")

	tests := []struct {
		propName    string
		isPrimitive bool
		isList      bool
		isMap       bool
		isComplex   bool
	}{
		{"BucketName", true, false, false, false},
		{"Tags", false, true, false, false},
		{"CorsConfiguration", false, false, false, true},
	}

	for _, tt := range tests {
		t.Run(tt.propName, func(t *testing.T) {
			prop := bucket.GetProperty(tt.propName)
			if prop == nil {
				t.Fatalf("property %s not found", tt.propName)
			}
			if prop.IsPrimitive() != tt.isPrimitive {
				t.Errorf("IsPrimitive() = %v, want %v", prop.IsPrimitive(), tt.isPrimitive)
			}
			if prop.IsList() != tt.isList {
				t.Errorf("IsList() = %v, want %v", prop.IsList(), tt.isList)
			}
			if prop.IsMap() != tt.isMap {
				t.Errorf("IsMap() = %v, want %v", prop.IsMap(), tt.isMap)
			}
			if prop.IsComplex() != tt.isComplex {
				t.Errorf("IsComplex() = %v, want %v", prop.IsComplex(), tt.isComplex)
			}
		})
	}
}

func TestPropertyType_GetRequiredProperties(t *testing.T) {
	s := loadTestSpec(t)
	tag := s.GetPropertyType("AWS::S3::Bucket.Tag")
	if tag == nil {
		t.Fatal("expected Tag property type")
	}

	required := tag.GetRequiredProperties()
	if len(required) != 2 {
		t.Errorf("expected 2 required properties, got %d", len(required))
	}
}

func TestGetPropertyTypeForResource(t *testing.T) {
	result := spec.GetPropertyTypeForResource("AWS::S3::Bucket", "CorsConfiguration")
	expected := "AWS::S3::Bucket.CorsConfiguration"
	if result != expected {
		t.Errorf("got %s, want %s", result, expected)
	}
}

func TestParsePropertyTypeName(t *testing.T) {
	tests := []struct {
		input        string
		wantResource string
		wantProperty string
	}{
		{"AWS::S3::Bucket.CorsConfiguration", "AWS::S3::Bucket", "CorsConfiguration"},
		{"AWS::S3::Bucket.Tag", "AWS::S3::Bucket", "Tag"},
		{"AWS::S3::Bucket", "AWS::S3::Bucket", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			gotResource, gotProperty := spec.ParsePropertyTypeName(tt.input)
			if gotResource != tt.wantResource || gotProperty != tt.wantProperty {
				t.Errorf("ParsePropertyTypeName(%q) = (%q, %q), want (%q, %q)",
					tt.input, gotResource, gotProperty, tt.wantResource, tt.wantProperty)
			}
		})
	}
}

func TestLoadSpec(t *testing.T) {
	// Create a temp file with test spec
	tmpDir := t.TempDir()
	specPath := filepath.Join(tmpDir, "spec.json")
	if err := os.WriteFile(specPath, []byte(testSpecJSON), 0644); err != nil {
		t.Fatalf("failed to write test spec: %v", err)
	}

	s, err := spec.LoadSpec(specPath)
	if err != nil {
		t.Fatalf("LoadSpec failed: %v", err)
	}

	if s.ResourceSpecificationVersion != "1.0.0" {
		t.Errorf("expected version 1.0.0, got %s", s.ResourceSpecificationVersion)
	}
	if len(s.ResourceTypes) != 2 {
		t.Errorf("expected 2 resource types, got %d", len(s.ResourceTypes))
	}
}

func TestLoadSpec_NotFound(t *testing.T) {
	_, err := spec.LoadSpec("/nonexistent/path/spec.json")
	if err == nil {
		t.Error("expected error for non-existent file")
	}
}

func TestLoadSpec_InvalidJSON(t *testing.T) {
	tmpDir := t.TempDir()
	specPath := filepath.Join(tmpDir, "invalid.json")
	if err := os.WriteFile(specPath, []byte("not valid json"), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	_, err := spec.LoadSpec(specPath)
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}
