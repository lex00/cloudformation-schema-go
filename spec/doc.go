// Package spec provides types for the CloudFormation Resource Specification.
//
// It includes functions to download, cache, and query the specification:
//
//	cfSpec, err := spec.FetchSpec(nil)
//	bucket := cfSpec.GetResourceType("AWS::S3::Bucket")
//	required := bucket.GetRequiredProperties()
package spec
