// Package enums provides CloudFormation enum constants and validation.
//
// Constants are generated from aws-sdk-go-v2 service types:
//
//	runtime := enums.LambdaRuntimePython312
//	arch := enums.LambdaArchitectureArm64
//
// Validation functions check if values are valid for an enum:
//
//	allowed := enums.GetAllowedValues("lambda", "Runtime")
//	valid := enums.IsValidValue("lambda", "Runtime", "python3.12")
//
// Discovery functions list available services and enums:
//
//	services := enums.Services()           // all services with enums
//	names := enums.GetEnumNames("lambda")  // all enum names for a service
//
// Regenerate from the latest SDK:
//
//	go generate ./enums/...
package enums
