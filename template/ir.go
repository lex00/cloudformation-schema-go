package template

// Property represents a resource property key-value pair.
type Property struct {
	Name  string // Original CloudFormation name (e.g., "BucketName")
	Value any    // Parsed value (may contain *Intrinsic)
}

// Parameter represents a CloudFormation parameter.
type Parameter struct {
	LogicalID             string
	Type                  string
	Description           string
	Default               any
	AllowedValues         []any
	AllowedPattern        string
	MinLength             *int
	MaxLength             *int
	MinValue              *float64
	MaxValue              *float64
	ConstraintDescription string
	NoEcho                bool
}

// Resource represents a CloudFormation resource.
type Resource struct {
	LogicalID           string
	ResourceType        string // e.g., "AWS::S3::Bucket"
	Properties          map[string]*Property
	DependsOn           []string
	Condition           string
	DeletionPolicy      string
	UpdateReplacePolicy string
	Metadata            map[string]any
}

// Service returns the AWS service name (e.g., "S3" from "AWS::S3::Bucket").
func (r *Resource) Service() string {
	parts := splitResourceType(r.ResourceType)
	if len(parts) >= 2 {
		return parts[1]
	}
	return ""
}

// TypeName returns the resource type name (e.g., "Bucket" from "AWS::S3::Bucket").
func (r *Resource) TypeName() string {
	parts := splitResourceType(r.ResourceType)
	if len(parts) >= 3 {
		return parts[2]
	}
	return ""
}

func splitResourceType(rt string) []string {
	var parts []string
	start := 0
	for i := 0; i < len(rt); i++ {
		if rt[i] == ':' && i+1 < len(rt) && rt[i+1] == ':' {
			parts = append(parts, rt[start:i])
			start = i + 2
			i++
		}
	}
	if start < len(rt) {
		parts = append(parts, rt[start:])
	}
	return parts
}

// Output represents a CloudFormation output.
type Output struct {
	LogicalID   string
	Value       any
	Description string
	ExportName  any // May be string or *Intrinsic
	Condition   string
}

// Mapping represents a CloudFormation mapping table.
type Mapping struct {
	LogicalID string
	MapData   map[string]map[string]any
}

// Condition represents a CloudFormation condition.
type Condition struct {
	LogicalID  string
	Expression any // Usually an *Intrinsic
}

// Template represents a complete parsed CloudFormation template.
type Template struct {
	Description              string
	AWSTemplateFormatVersion string
	Parameters               map[string]*Parameter
	Mappings                 map[string]*Mapping
	Conditions               map[string]*Condition
	Resources                map[string]*Resource
	Outputs                  map[string]*Output
	SourceFile               string
	ReferenceGraph           map[string][]string // resource -> list of resources it references
}

// NewTemplate creates a new empty template.
func NewTemplate() *Template {
	return &Template{
		AWSTemplateFormatVersion: "2010-09-09",
		Parameters:               make(map[string]*Parameter),
		Mappings:                 make(map[string]*Mapping),
		Conditions:               make(map[string]*Condition),
		Resources:                make(map[string]*Resource),
		Outputs:                  make(map[string]*Output),
		ReferenceGraph:           make(map[string][]string),
	}
}
