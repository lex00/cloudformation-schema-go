package enums_test

import (
	"testing"

	"github.com/lex00/cloudformation-schema-go/enums"
)

func TestGetAllowedValues_Lambda(t *testing.T) {
	values := enums.GetAllowedValues("lambda", "Runtime")
	if values == nil {
		t.Fatal("expected Runtime values, got nil")
	}
	if len(values) < 10 {
		t.Errorf("expected at least 10 Runtime values, got %d", len(values))
	}

	// Check that python3.12 is in the list
	found := false
	for _, v := range values {
		if v == "python3.12" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected python3.12 in Runtime values")
	}
}

func TestGetAllowedValues_NotFound(t *testing.T) {
	values := enums.GetAllowedValues("nonexistent", "Runtime")
	if values != nil {
		t.Errorf("expected nil for nonexistent service, got %v", values)
	}

	values = enums.GetAllowedValues("lambda", "NonexistentEnum")
	if values != nil {
		t.Errorf("expected nil for nonexistent enum, got %v", values)
	}
}

func TestIsValidValue(t *testing.T) {
	tests := []struct {
		service  string
		enumName string
		value    string
		want     bool
	}{
		{"lambda", "Runtime", "python3.12", true},
		{"lambda", "Runtime", "invalid-runtime", false},
		{"lambda", "Architecture", "arm64", true},
		{"lambda", "Architecture", "x86_64", true},
		{"lambda", "Architecture", "ppc64", false},
		{"s3", "StorageClass", "STANDARD", true},
		{"s3", "StorageClass", "INVALID", false},
		{"nonexistent", "Runtime", "python3.12", false},
	}

	for _, tt := range tests {
		t.Run(tt.service+"/"+tt.enumName+"/"+tt.value, func(t *testing.T) {
			got := enums.IsValidValue(tt.service, tt.enumName, tt.value)
			if got != tt.want {
				t.Errorf("IsValidValue(%q, %q, %q) = %v, want %v",
					tt.service, tt.enumName, tt.value, got, tt.want)
			}
		})
	}
}

func TestGetEnumNames(t *testing.T) {
	names := enums.GetEnumNames("lambda")
	if names == nil {
		t.Fatal("expected enum names, got nil")
	}

	expected := map[string]bool{
		"Runtime":      true,
		"Architecture": true,
		"PackageType":  true,
	}

	for _, name := range names {
		if !expected[name] {
			t.Errorf("unexpected enum name: %s", name)
		}
		delete(expected, name)
	}

	for name := range expected {
		t.Errorf("missing enum name: %s", name)
	}
}

func TestGetEnumNames_NotFound(t *testing.T) {
	names := enums.GetEnumNames("nonexistent")
	if names != nil {
		t.Errorf("expected nil for nonexistent service, got %v", names)
	}
}

func TestServices(t *testing.T) {
	services := enums.Services()
	if len(services) < 5 {
		t.Errorf("expected at least 5 services, got %d", len(services))
	}

	// Check that lambda is in the list
	found := false
	for _, s := range services {
		if s == "lambda" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected lambda in services list")
	}
}

func TestConstants(t *testing.T) {
	// Verify constants are accessible and have correct values
	if enums.LambdaRuntimePython312 != "python3.12" {
		t.Errorf("LambdaRuntimePython312 = %q, want %q",
			enums.LambdaRuntimePython312, "python3.12")
	}

	if enums.LambdaArchitectureArm64 != "arm64" {
		t.Errorf("LambdaArchitectureArm64 = %q, want %q",
			enums.LambdaArchitectureArm64, "arm64")
	}

	if enums.S3StorageClassStandard != "STANDARD" {
		t.Errorf("S3StorageClassStandard = %q, want %q",
			enums.S3StorageClassStandard, "STANDARD")
	}
}

func TestEC2VolumeType(t *testing.T) {
	values := enums.GetAllowedValues("ec2", "VolumeType")
	if values == nil {
		t.Fatal("expected VolumeType values, got nil")
	}

	// Check for common volume types
	commonTypes := []string{"gp2", "gp3", "io1", "io2", "standard"}
	for _, ct := range commonTypes {
		if !enums.IsValidValue("ec2", "VolumeType", ct) {
			t.Errorf("expected %s to be a valid VolumeType", ct)
		}
	}
}

func TestGetEnumForProperty(t *testing.T) {
	tests := []struct {
		service      string
		propertyName string
		want         string
	}{
		{"lambda", "Runtime", "Runtime"},
		{"lambda", "PackageType", "PackageType"},
		{"s3", "StorageClass", "StorageClass"},
		{"s3", "AccessControl", "BucketCannedACL"},
		{"s3", "SSEAlgorithm", "ServerSideEncryption"},
		{"dynamodb", "BillingMode", "BillingMode"},
		{"lambda", "NonexistentProperty", ""},
		{"nonexistent", "Runtime", ""},
	}

	for _, tt := range tests {
		t.Run(tt.service+"/"+tt.propertyName, func(t *testing.T) {
			got := enums.GetEnumForProperty(tt.service, tt.propertyName)
			if got != tt.want {
				t.Errorf("GetEnumForProperty(%q, %q) = %q, want %q",
					tt.service, tt.propertyName, got, tt.want)
			}
		})
	}
}

func TestDynamoDBEnums(t *testing.T) {
	// BillingMode
	if !enums.IsValidValue("dynamodb", "BillingMode", "PROVISIONED") {
		t.Error("expected PROVISIONED to be a valid BillingMode")
	}
	if !enums.IsValidValue("dynamodb", "BillingMode", "PAY_PER_REQUEST") {
		t.Error("expected PAY_PER_REQUEST to be a valid BillingMode")
	}

	// StreamViewType
	values := enums.GetAllowedValues("dynamodb", "StreamViewType")
	if values == nil {
		t.Fatal("expected StreamViewType values, got nil")
	}
}
