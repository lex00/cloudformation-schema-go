// enumgen generates enum constants from enums.json (extracted from botocore).
//
// Usage:
//
//	go run ./cmd/enumgen
//
// This will generate enums/*.go files with constants and lookup functions.
// The enums.json file must be generated first by running:
//
//	python scripts/extract_enums.py
package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"go/format"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/template"
)

var outputDir = flag.String("output", "enums", "output directory for generated files")

// EnumsJSON is the structure of enums.json
type EnumsJSON struct {
	Services map[string]map[string]EnumType `json:"services"`
}

// EnumType represents an enum type in the JSON
type EnumType struct {
	Name   string      `json:"name"`
	Values []EnumValue `json:"values"`
}

// EnumValue represents a single enum value
type EnumValue struct {
	Value  string `json:"value"`
	GoName string `json:"goName"`
	PyName string `json:"pyName"`
}

// ServiceData represents data for generating a service enum file.
type ServiceData struct {
	Service     string // Original service name (e.g., "acm-pca") - used for file names
	ServiceName string // PascalCase service name for Go (e.g., "AcmPca") - used for function names
	ServiceID   string // Lowercase service name without hyphens (e.g., "acmpca") - used for variable names
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

	// Load enums.json
	// When run via go generate from enums/, outputDir is "enums" but we're already in enums/
	// so look for enums.json in the current directory first, then adjust outputDir
	actualOutputDir := *outputDir
	enumsPath := filepath.Join(actualOutputDir, "enums.json")
	if _, err := os.Stat(enumsPath); os.IsNotExist(err) {
		// Try looking in current directory (for go generate case)
		if _, err := os.Stat("enums.json"); err == nil {
			enumsPath = "enums.json"
			actualOutputDir = "." // Write to current directory
		}
	}
	data, err := os.ReadFile(enumsPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to read %s: %v\n", enumsPath, err)
		fmt.Fprintf(os.Stderr, "hint: run 'python scripts/extract_enums.py' first\n")
		os.Exit(1)
	}

	var enums EnumsJSON
	if err := json.Unmarshal(data, &enums); err != nil {
		fmt.Fprintf(os.Stderr, "failed to parse enums.json: %v\n", err)
		os.Exit(1)
	}

	var allServices []string
	serviceEnumNames := make(map[string][]string)

	// Get sorted list of all services for deterministic output
	var services []string
	for service := range enums.Services {
		services = append(services, service)
	}
	sort.Strings(services)

	for _, service := range services {
		serviceEnums := enums.Services[service]

		svcData := ServiceData{
			Service:     service,
			ServiceName: capitalize(service),
			ServiceID:   strings.ReplaceAll(service, "-", ""),
		}

		// Get sorted list of enum names for deterministic output
		var enumNames []string
		for enumName := range serviceEnums {
			enumNames = append(enumNames, enumName)
		}
		sort.Strings(enumNames)

		for _, enumName := range enumNames {
			enumType := serviceEnums[enumName]

			info := EnumInfo{Name: enumName}
			for _, v := range enumType.Values {
				// Sanitize goName by removing hyphens (service names like "acm-pca" have hyphens)
				goName := strings.ReplaceAll(v.GoName, "-", "")
				info.Constants = append(info.Constants, ConstantInfo{
					Name:  goName,
					Value: v.Value,
				})
			}
			svcData.Enums = append(svcData.Enums, info)
			serviceEnumNames[service] = append(serviceEnumNames[service], enumName)
		}

		if len(svcData.Enums) == 0 {
			continue
		}

		allServices = append(allServices, service)

		if err := generateServiceFile(svcData, actualOutputDir); err != nil {
			fmt.Fprintf(os.Stderr, "failed to generate %s: %v\n", service, err)
			os.Exit(1)
		}
		fmt.Printf("Generated %s/%s.go\n", actualOutputDir, service)
	}

	if err := generateLookupFile(allServices, serviceEnumNames, enums, actualOutputDir); err != nil {
		fmt.Fprintf(os.Stderr, "failed to generate lookup.go: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Generated %s/lookup.go\n", actualOutputDir)
}

func capitalize(s string) string {
	if s == "" {
		return s
	}

	// Handle hyphenated names by converting to PascalCase
	if strings.Contains(s, "-") {
		parts := strings.Split(s, "-")
		var result strings.Builder
		for _, part := range parts {
			result.WriteString(capitalize(part))
		}
		return result.String()
	}

	// Handle common acronyms and short names
	switch s {
	case "ec2":
		return "Ec2"
	case "ecs":
		return "Ecs"
	case "rds":
		return "Rds"
	case "s3":
		return "S3"
	case "acm":
		return "Acm"
	case "sns":
		return "Sns"
	case "sqs":
		return "Sqs"
	case "iam":
		return "Iam"
	case "kms":
		return "Kms"
	case "ses":
		return "Ses"
	case "waf":
		return "Waf"
	case "efs":
		return "Efs"
	case "emr":
		return "Emr"
	case "elb":
		return "Elb"
	case "ebs":
		return "Ebs"
	case "vpc":
		return "Vpc"
	case "iot":
		return "Iot"
	case "mq":
		return "Mq"
	case "ds":
		return "Ds"
	case "ecr":
		return "Ecr"
	case "eks":
		return "Eks"
	case "ram":
		return "Ram"
	case "fsx":
		return "Fsx"
	case "dax":
		return "Dax"
	case "dms":
		return "Dms"
	case "dlm":
		return "Dlm"
	case "fms":
		return "Fms"
	case "ssm":
		return "Ssm"
	case "sts":
		return "Sts"
	case "amp":
		return "Amp"
	case "aps":
		return "Aps"
	case "aoss":
		return "Aoss"
	case "b2bi":
		return "B2bi"
	case "pca":
		return "Pca"
	case "sdk":
		return "Sdk"
	case "idp":
		return "Idp"
	default:
		return strings.ToUpper(s[:1]) + s[1:]
	}
}

const serviceTemplate = `// Code generated by enumgen from enums.json. DO NOT EDIT.

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
var {{$.ServiceID}}{{.Name}}Values = []string{
{{- range .Constants}}
	"{{.Value}}",
{{- end}}
}
{{end}}

func get{{.ServiceName}}Enum(name string) []string {
	switch name {
{{- range .Enums}}
	case "{{.Name}}":
		return {{$.ServiceID}}{{.Name}}Values
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

func generateServiceFile(data ServiceData, outputDir string) error {
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

	path := filepath.Join(outputDir, data.Service+".go")
	return os.WriteFile(path, formatted, 0644)
}

const lookupTemplate = `// Code generated by enumgen from enums.json. DO NOT EDIT.

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

func generateLookupFile(services []string, enumNames map[string][]string, allEnums EnumsJSON, outputDir string) error {
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

	path := filepath.Join(outputDir, "lookup.go")
	return os.WriteFile(path, formatted, 0644)
}
