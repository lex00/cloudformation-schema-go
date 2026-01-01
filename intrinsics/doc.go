// Package intrinsics provides CloudFormation intrinsic function types.
//
// All types implement json.Marshaler to serialize to CloudFormation syntax:
//
//	Ref{"MyBucket"}           → {"Ref": "MyBucket"}
//	GetAtt{"MyRole", "Arn"}   → {"Fn::GetAtt": ["MyRole", "Arn"]}
//	Sub{"${AWS::Region}-x"}   → {"Fn::Sub": "${AWS::Region}-x"}
//
// Pseudo-parameters are provided as pre-defined Ref values:
//
//	AWS_REGION      → {"Ref": "AWS::Region"}
//	AWS_ACCOUNT_ID  → {"Ref": "AWS::AccountId"}
package intrinsics
