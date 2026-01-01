// enumgen generates enum constants from aws-sdk-go-v2 service packages.
//
// Usage:
//
//	go run ./cmd/enumgen
//
// This will generate enums/*.go files with constants and lookup functions.
package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/format"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"text/template"

	// Import service types packages
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	ecstypes "github.com/aws/aws-sdk-go-v2/service/ecs/types"
	lambdatypes "github.com/aws/aws-sdk-go-v2/service/lambda/types"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"
)

var outputDir = flag.String("output", "enums", "output directory for generated files")

// ServiceEnums defines which enums to extract from each service.
type ServiceEnums struct {
	Service    string
	CFNService string // CloudFormation service name if different
	EnumTypes  []EnumExtractor
}

// EnumExtractor extracts values from a typed enum.
type EnumExtractor struct {
	Name   string
	Values func() []string
}

// priorityServices lists services to generate enums for.
var priorityServices = []ServiceEnums{
	{
		Service: "lambda",
		EnumTypes: []EnumExtractor{
			{"Runtime", lambdaRuntimeValues},
			{"Architecture", lambdaArchitectureValues},
			{"PackageType", lambdaPackageTypeValues},
		},
	},
	{
		Service: "ec2",
		EnumTypes: []EnumExtractor{
			{"InstanceType", ec2InstanceTypeValues},
			{"VolumeType", ec2VolumeTypeValues},
		},
	},
	{
		Service: "ecs",
		EnumTypes: []EnumExtractor{
			{"LaunchType", ecsLaunchTypeValues},
			{"SchedulingStrategy", ecsSchedulingStrategyValues},
		},
	},
	{
		Service: "rds",
		EnumTypes: []EnumExtractor{
			{"DBInstanceClass", rdsDBInstanceClassValues},
			{"EngineType", rdsEngineTypeValues},
		},
	},
	{
		Service: "s3",
		EnumTypes: []EnumExtractor{
			{"StorageClass", s3StorageClassValues},
			{"ObjectCannedACL", s3ObjectCannedACLValues},
			{"BucketCannedACL", s3BucketCannedACLValues},
		},
	},
	{
		Service: "dynamodb",
		EnumTypes: []EnumExtractor{
			{"BillingMode", dynamodbBillingModeValues},
			{"StreamViewType", dynamodbStreamViewTypeValues},
			{"TableClass", dynamodbTableClassValues},
		},
	},
}

// Wrapper functions to get string values from typed enums
func lambdaRuntimeValues() []string {
	vals := lambdatypes.Runtime("").Values()
	result := make([]string, len(vals))
	for i, v := range vals {
		result[i] = string(v)
	}
	return result
}

func lambdaArchitectureValues() []string {
	vals := lambdatypes.Architecture("").Values()
	result := make([]string, len(vals))
	for i, v := range vals {
		result[i] = string(v)
	}
	return result
}

func lambdaPackageTypeValues() []string {
	vals := lambdatypes.PackageType("").Values()
	result := make([]string, len(vals))
	for i, v := range vals {
		result[i] = string(v)
	}
	return result
}

func ec2InstanceTypeValues() []string {
	vals := ec2types.InstanceType("").Values()
	result := make([]string, len(vals))
	for i, v := range vals {
		result[i] = string(v)
	}
	return result
}

func ec2VolumeTypeValues() []string {
	vals := ec2types.VolumeType("").Values()
	result := make([]string, len(vals))
	for i, v := range vals {
		result[i] = string(v)
	}
	return result
}

func ecsLaunchTypeValues() []string {
	vals := ecstypes.LaunchType("").Values()
	result := make([]string, len(vals))
	for i, v := range vals {
		result[i] = string(v)
	}
	return result
}

func ecsSchedulingStrategyValues() []string {
	vals := ecstypes.SchedulingStrategy("").Values()
	result := make([]string, len(vals))
	for i, v := range vals {
		result[i] = string(v)
	}
	return result
}

func rdsDBInstanceClassValues() []string {
	// RDS doesn't have a direct Values() enum, use common ones
	return []string{
		"db.t3.micro", "db.t3.small", "db.t3.medium", "db.t3.large",
		"db.t3.xlarge", "db.t3.2xlarge",
		"db.r5.large", "db.r5.xlarge", "db.r5.2xlarge", "db.r5.4xlarge",
		"db.m5.large", "db.m5.xlarge", "db.m5.2xlarge", "db.m5.4xlarge",
	}
}

func rdsEngineTypeValues() []string {
	// Common RDS engine types
	return []string{
		"mysql", "postgres", "mariadb", "oracle-ee", "oracle-se2",
		"sqlserver-ee", "sqlserver-se", "sqlserver-ex", "sqlserver-web",
		"aurora", "aurora-mysql", "aurora-postgresql",
	}
}

func s3StorageClassValues() []string {
	vals := s3types.StorageClass("").Values()
	result := make([]string, len(vals))
	for i, v := range vals {
		result[i] = string(v)
	}
	return result
}

func s3ObjectCannedACLValues() []string {
	vals := s3types.ObjectCannedACL("").Values()
	result := make([]string, len(vals))
	for i, v := range vals {
		result[i] = string(v)
	}
	return result
}

func s3BucketCannedACLValues() []string {
	vals := s3types.BucketCannedACL("").Values()
	result := make([]string, len(vals))
	for i, v := range vals {
		result[i] = string(v)
	}
	return result
}

func dynamodbBillingModeValues() []string {
	vals := types.BillingMode("").Values()
	result := make([]string, len(vals))
	for i, v := range vals {
		result[i] = string(v)
	}
	return result
}

func dynamodbStreamViewTypeValues() []string {
	vals := types.StreamViewType("").Values()
	result := make([]string, len(vals))
	for i, v := range vals {
		result[i] = string(v)
	}
	return result
}

func dynamodbTableClassValues() []string {
	vals := types.TableClass("").Values()
	result := make([]string, len(vals))
	for i, v := range vals {
		result[i] = string(v)
	}
	return result
}

// EnumData represents data for generating a service enum file.
type EnumData struct {
	Service     string
	ServiceName string // Capitalized service name
	Enums       []EnumInfo
}

// EnumInfo represents a single enum type.
type EnumInfo struct {
	Name      string
	Constants []ConstantInfo
}

// ConstantInfo represents a single constant.
type ConstantInfo struct {
	Name  string
	Value string
}

func main() {
	flag.Parse()

	if err := os.MkdirAll(*outputDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "failed to create output directory: %v\n", err)
		os.Exit(1)
	}

	var allServices []string
	serviceEnumNames := make(map[string][]string)

	for _, svc := range priorityServices {
		data := extractServiceEnums(svc)
		if len(data.Enums) == 0 {
			continue
		}

		allServices = append(allServices, svc.Service)
		for _, e := range data.Enums {
			serviceEnumNames[svc.Service] = append(serviceEnumNames[svc.Service], e.Name)
		}

		if err := generateServiceFile(data); err != nil {
			fmt.Fprintf(os.Stderr, "failed to generate %s: %v\n", svc.Service, err)
			os.Exit(1)
		}
		fmt.Printf("Generated enums/%s.go\n", svc.Service)
	}

	if err := generateLookupFile(allServices, serviceEnumNames); err != nil {
		fmt.Fprintf(os.Stderr, "failed to generate lookup.go: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Generated enums/lookup.go\n")
}

func extractServiceEnums(svc ServiceEnums) EnumData {
	data := EnumData{
		Service:     svc.Service,
		ServiceName: capitalize(svc.Service),
	}

	for _, extractor := range svc.EnumTypes {
		values := extractor.Values()
		if len(values) == 0 {
			continue
		}

		info := EnumInfo{Name: extractor.Name}
		for _, v := range values {
			constName := toConstName(svc.Service, extractor.Name, v)
			info.Constants = append(info.Constants, ConstantInfo{
				Name:  constName,
				Value: v,
			})
		}
		data.Enums = append(data.Enums, info)
	}

	return data
}

func toConstName(service, enumName, value string) string {
	// Convert value to valid Go identifier
	// e.g., "python3.12" -> "Python312", "t3.micro" -> "T3Micro"
	name := capitalize(service) + enumName

	// Replace dots with nothing, capitalize after
	parts := strings.Split(value, ".")
	for _, part := range parts {
		// Replace hyphens and underscores, capitalize
		subparts := strings.FieldsFunc(part, func(r rune) bool {
			return r == '-' || r == '_'
		})
		for _, sp := range subparts {
			name += capitalize(sp)
		}
	}

	return name
}

func capitalize(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

const serviceTemplate = `// Code generated by enumgen. DO NOT EDIT.

package enums

// {{.ServiceName}} enum constants
const (
{{- range .Enums}}
	// {{.Name}} values
{{- range .Constants}}
	{{.Name}} = "{{.Value}}"
{{- end}}
{{end}}
)

{{range .Enums}}
var {{$.Service}}{{.Name}}Values = []string{
{{- range .Constants}}
	"{{.Value}}",
{{- end}}
}
{{end}}

func get{{.ServiceName}}Enum(name string) []string {
	switch name {
{{- range .Enums}}
	case "{{.Name}}":
		return {{$.Service}}{{.Name}}Values
{{- end}}
	}
	return nil
}

func get{{.ServiceName}}EnumNames() []string {
	return []string{
{{- range .Enums}}
		"{{.Name}}",
{{- end}}
	}
}
`

func generateServiceFile(data EnumData) error {
	tmpl, err := template.New("service").Parse(serviceTemplate)
	if err != nil {
		return err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return err
	}

	formatted, err := format.Source(buf.Bytes())
	if err != nil {
		// Write unformatted for debugging
		formatted = buf.Bytes()
	}

	path := filepath.Join(*outputDir, data.Service+".go")
	return os.WriteFile(path, formatted, 0644)
}

const lookupTemplate = `// Code generated by enumgen. DO NOT EDIT.

package enums

// GetAllowedValues returns valid values for a service enum.
// Returns nil if the service or enum is not found.
func GetAllowedValues(service, enumName string) []string {
	switch service {
{{- range .Services}}
	case "{{.}}":
		return get{{. | capitalize}}Enum(enumName)
{{- end}}
	}
	return nil
}

// IsValidValue checks if a value is valid for an enum.
func IsValidValue(service, enumName, value string) bool {
	values := GetAllowedValues(service, enumName)
	for _, v := range values {
		if v == value {
			return true
		}
	}
	return false
}

// GetEnumNames returns all enum names for a service.
// Returns nil if the service is not found.
func GetEnumNames(service string) []string {
	switch service {
{{- range .Services}}
	case "{{.}}":
		return get{{. | capitalize}}EnumNames()
{{- end}}
	}
	return nil
}

// Services returns the list of services with enums.
func Services() []string {
	return []string{
{{- range .Services}}
		"{{.}}",
{{- end}}
	}
}
`

type LookupData struct {
	Services []string
}

func generateLookupFile(services []string, enumNames map[string][]string) error {
	sort.Strings(services)

	funcMap := template.FuncMap{
		"capitalize": capitalize,
	}

	tmpl, err := template.New("lookup").Funcs(funcMap).Parse(lookupTemplate)
	if err != nil {
		return err
	}

	data := LookupData{Services: services}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return err
	}

	formatted, err := format.Source(buf.Bytes())
	if err != nil {
		formatted = buf.Bytes()
	}

	path := filepath.Join(*outputDir, "lookup.go")
	return os.WriteFile(path, formatted, 0644)
}

// Ensure we're actually using reflect to avoid unused import
var _ = reflect.TypeOf(nil)
