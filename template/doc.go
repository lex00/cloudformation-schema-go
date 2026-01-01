// Package template provides CloudFormation template parsing.
//
// Templates are parsed into an intermediate representation (IR) that
// preserves intrinsic functions and builds a reference graph:
//
//	tmpl, err := template.ParseTemplate("template.yaml")
//	for name, resource := range tmpl.Resources {
//	    fmt.Printf("%s: %s\n", name, resource.ResourceType)
//	}
//
// Both YAML and JSON formats are supported, including short-form
// intrinsic syntax (!Ref, !Sub, etc.) and long-form ({"Ref": ...}).
package template
