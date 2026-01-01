# cloudformation-schema-go

Go types for CloudFormation specifications, intrinsic functions, template parsing, and enum validation.

Shared foundation for:
- [wetwire-aws](https://github.com/lex00/wetwire-aws) - Declarative IaC in Go
- [cfn-lint-go](https://github.com/lex00/cfn-lint-go) - CloudFormation linter in Go

## Packages

### spec/

CloudFormation Resource Specification types and fetching.

```go
import "github.com/lex00/cloudformation-schema-go/spec"

// Download and cache the CF spec
cfSpec, err := spec.FetchSpec()

// Look up resource types
bucket := cfSpec.GetResourceType("AWS::S3::Bucket")
required := bucket.GetRequiredProperties()
```

### intrinsics/

CloudFormation intrinsic functions with JSON marshaling.

```go
import "github.com/lex00/cloudformation-schema-go/intrinsics"

// Intrinsic functions
ref := intrinsics.Ref{LogicalName: "MyBucket"}
getAtt := intrinsics.GetAtt{LogicalName: "MyRole", Attribute: "Arn"}
sub := intrinsics.Sub{String: "${AWS::Region}-bucket"}

// Pseudo-parameters
region := intrinsics.AWS_REGION
accountID := intrinsics.AWS_ACCOUNT_ID
```

### template/

Parse CloudFormation YAML/JSON templates to intermediate representation.

```go
import "github.com/lex00/cloudformation-schema-go/template"

// Parse a template file
tmpl, err := template.ParseTemplate("template.yaml")

// Access resources
for name, resource := range tmpl.Resources {
    fmt.Printf("%s: %s\n", name, resource.ResourceType)
}

// Reference graph for dependency analysis
deps := tmpl.ReferenceGraph["MyFunction"]  // ["MyRole", "MyBucket"]
```

### enums/

Enum constants and validation, generated from aws-sdk-go-v2.

```go
import "github.com/lex00/cloudformation-schema-go/enums"

// Use constants for IDE autocomplete
runtime := enums.LambdaRuntimePython312

// Validate enum values
allowed := enums.GetAllowedValues("lambda", "Runtime")
valid := enums.IsValidValue("lambda", "Runtime", "python3.12")
```

## Installation

```bash
go get github.com/lex00/cloudformation-schema-go
```

## Regenerating Enums

Enum constants are generated from aws-sdk-go-v2:

```bash
go generate ./enums/...
```

## License

MIT
