package enums

// PropertyEnumMapping maps (service, propertyName) to enum type name.
// This helps importers and linters know which properties accept enum values.
// Property names are in PascalCase as used in CloudFormation.
//
// Example usage:
//
//	enumName := enums.GetEnumForProperty("lambda", "Runtime")
//	if enumName != "" {
//	    values := enums.GetAllowedValues("lambda", enumName)
//	}
var PropertyEnumMapping = map[string]map[string]string{
	"lambda": {
		"Runtime":      "Runtime",
		"PackageType":  "PackageType",
		"Architecture": "Architecture",
	},
	"ec2": {
		"VolumeType": "VolumeType",
	},
	"ecs": {
		"LaunchType":         "LaunchType",
		"SchedulingStrategy": "SchedulingStrategy",
		"NetworkMode":        "NetworkMode",
	},
	"s3": {
		"StorageClass":    "StorageClass",
		"AccessControl":   "BucketCannedACL",
		"SSEAlgorithm":    "ServerSideEncryption",
		"Mode":            "ObjectLockRetentionMode",
		"Status":          "BucketVersioningStatus",
		"Protocol":        "Protocol",
		"ObjectCannedAcl": "ObjectCannedACL",
	},
	"dynamodb": {
		"BillingMode":    "BillingMode",
		"StreamViewType": "StreamViewType",
		"TableClass":     "TableClass",
	},
	"apigateway": {
		"IntegrationType": "IntegrationType",
	},
	"elbv2": {
		"Protocol":   "ProtocolEnum",
		"TargetType": "TargetTypeEnum",
	},
	"logs": {
		"LogGroupClass": "LogGroupClass",
	},
	"acm": {
		"ValidationMethod":  "ValidationMethod",
		"CertificateStatus": "CertificateStatus",
	},
	"events": {
		"State": "RuleState",
	},
}

// GetEnumForProperty returns the enum type name for a service property.
// Returns empty string if the property doesn't have an enum mapping.
func GetEnumForProperty(service, propertyName string) string {
	if svc, ok := PropertyEnumMapping[service]; ok {
		return svc[propertyName]
	}
	return ""
}
