# cloudformation-schema-go

Go types for CloudFormation specifications, intrinsic functions, template parsing, enum validation, and code generation utilities.

Shared foundation for:
- [wetwire-aws](https://github.com/lex00/wetwire-aws) - Declarative IaC in Go
- [cfn-lint-go](https://github.com/lex00/cfn-lint-go) - CloudFormation linter in Go

## Packages

### spec/

CloudFormation Resource Specification types and fetching.

```go
import "github.com/lex00/cloudformation-schema-go/spec"

// Download and cache the CF spec (use nil for defaults)
cfSpec, err := spec.FetchSpec(nil)

// Or with custom options
cfSpec, err := spec.FetchSpec(&spec.FetchOptions{
    CacheDir: "/custom/cache",
    MaxAge:   24 * time.Hour,
})

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

### codegen/

Utilities for code generation: case conversion, identifier sanitization, and topological sorting.

```go
import "github.com/lex00/cloudformation-schema-go/codegen"

// Case conversion
codegen.ToSnakeCase("BucketName")     // "bucket_name"
codegen.ToPascalCase("bucket_name")   // "BucketName"

// Go keyword handling
codegen.IsGoKeyword("type")           // true
codegen.IsGoKeyword("bucket")         // false

// Identifier sanitization (ensures valid Go identifiers)
codegen.SanitizeGoIdentifier("123start")   // "_123start"
codegen.SanitizeGoIdentifier("with-dash")  // "withdash"
codegen.SanitizeGoIdentifier("type")       // "type_"

// Topological sort for dependency ordering
nodes := []string{"A", "B", "C"}
deps := map[string][]string{"A": {"B"}, "B": {"C"}, "C": {}}
sorted := codegen.TopologicalSort(nodes, func(n string) []string {
    return deps[n]
})
// Result: ["C", "B", "A"] (dependencies first)
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

Apache 2.0 - See [LICENSE](LICENSE) and [NOTICE](NOTICE) for details.
