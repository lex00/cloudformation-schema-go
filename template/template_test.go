package template_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/lex00/cloudformation-schema-go/template"
)

const testYAMLTemplate = `AWSTemplateFormatVersion: "2010-09-09"
Description: Test template

Parameters:
  Environment:
    Type: String
    Default: dev
    AllowedValues:
      - dev
      - prod

Mappings:
  RegionMap:
    us-east-1:
      AMI: ami-12345678
    us-west-2:
      AMI: ami-87654321

Conditions:
  IsProd: !Equals [!Ref Environment, prod]

Resources:
  MyBucket:
    Type: AWS::S3::Bucket
    Properties:
      BucketName: !Sub "${Environment}-my-bucket"
      Tags:
        - Key: Environment
          Value: !Ref Environment
    Condition: IsProd
    DeletionPolicy: Retain

  MyRole:
    Type: AWS::IAM::Role
    Properties:
      RoleName: !Join ["-", [!Ref Environment, "role"]]
      AssumeRolePolicyDocument:
        Version: "2012-10-17"
        Statement:
          - Effect: Allow
            Principal:
              Service: lambda.amazonaws.com
            Action: sts:AssumeRole
    DependsOn: MyBucket

Outputs:
  BucketArn:
    Description: The ARN of the bucket
    Value: !GetAtt MyBucket.Arn
    Export:
      Name: !Sub "${AWS::StackName}-BucketArn"
`

const testJSONTemplate = `{
  "AWSTemplateFormatVersion": "2010-09-09",
  "Description": "JSON test template",
  "Resources": {
    "MyBucket": {
      "Type": "AWS::S3::Bucket",
      "Properties": {
        "BucketName": {"Fn::Sub": "${AWS::StackName}-bucket"}
      }
    }
  },
  "Outputs": {
    "BucketArn": {
      "Value": {"Fn::GetAtt": ["MyBucket", "Arn"]}
    }
  }
}`

func TestParseTemplateContent_YAML(t *testing.T) {
	tmpl, err := template.ParseTemplateContent([]byte(testYAMLTemplate), "test.yaml")
	if err != nil {
		t.Fatalf("failed to parse YAML template: %v", err)
	}

	// Check description
	if tmpl.Description != "Test template" {
		t.Errorf("expected description 'Test template', got %q", tmpl.Description)
	}

	// Check parameters
	if len(tmpl.Parameters) != 1 {
		t.Fatalf("expected 1 parameter, got %d", len(tmpl.Parameters))
	}
	env := tmpl.Parameters["Environment"]
	if env == nil {
		t.Fatal("expected Environment parameter")
	}
	if env.Type != "String" {
		t.Errorf("expected parameter type String, got %s", env.Type)
	}
	if env.Default != "dev" {
		t.Errorf("expected default 'dev', got %v", env.Default)
	}

	// Check mappings
	if len(tmpl.Mappings) != 1 {
		t.Fatalf("expected 1 mapping, got %d", len(tmpl.Mappings))
	}
	regionMap := tmpl.Mappings["RegionMap"]
	if regionMap == nil {
		t.Fatal("expected RegionMap mapping")
	}

	// Check conditions
	if len(tmpl.Conditions) != 1 {
		t.Fatalf("expected 1 condition, got %d", len(tmpl.Conditions))
	}
	isProd := tmpl.Conditions["IsProd"]
	if isProd == nil {
		t.Fatal("expected IsProd condition")
	}
	// The condition should be parsed as an Equals intrinsic
	intrinsic, ok := isProd.Expression.(*template.Intrinsic)
	if !ok {
		t.Fatalf("expected condition to be *Intrinsic, got %T", isProd.Expression)
	}
	if intrinsic.Type != template.IntrinsicEquals {
		t.Errorf("expected IntrinsicEquals, got %s", intrinsic.Type)
	}

	// Check resources
	if len(tmpl.Resources) != 2 {
		t.Fatalf("expected 2 resources, got %d", len(tmpl.Resources))
	}
	bucket := tmpl.Resources["MyBucket"]
	if bucket == nil {
		t.Fatal("expected MyBucket resource")
	}
	if bucket.ResourceType != "AWS::S3::Bucket" {
		t.Errorf("expected AWS::S3::Bucket, got %s", bucket.ResourceType)
	}
	if bucket.Service() != "S3" {
		t.Errorf("expected service S3, got %s", bucket.Service())
	}
	if bucket.TypeName() != "Bucket" {
		t.Errorf("expected type name Bucket, got %s", bucket.TypeName())
	}
	if bucket.Condition != "IsProd" {
		t.Errorf("expected condition IsProd, got %s", bucket.Condition)
	}
	if bucket.DeletionPolicy != "Retain" {
		t.Errorf("expected deletion policy Retain, got %s", bucket.DeletionPolicy)
	}

	role := tmpl.Resources["MyRole"]
	if role == nil {
		t.Fatal("expected MyRole resource")
	}
	if len(role.DependsOn) != 1 || role.DependsOn[0] != "MyBucket" {
		t.Errorf("expected DependsOn [MyBucket], got %v", role.DependsOn)
	}

	// Check outputs
	if len(tmpl.Outputs) != 1 {
		t.Fatalf("expected 1 output, got %d", len(tmpl.Outputs))
	}
	bucketArn := tmpl.Outputs["BucketArn"]
	if bucketArn == nil {
		t.Fatal("expected BucketArn output")
	}
	if bucketArn.Description != "The ARN of the bucket" {
		t.Errorf("expected description, got %q", bucketArn.Description)
	}
	outputValue, ok := bucketArn.Value.(*template.Intrinsic)
	if !ok {
		t.Fatalf("expected output value to be *Intrinsic, got %T", bucketArn.Value)
	}
	if outputValue.Type != template.IntrinsicGetAtt {
		t.Errorf("expected IntrinsicGetAtt, got %s", outputValue.Type)
	}

	// Check reference graph
	// MyBucket references Environment twice: once in Sub "${Environment}-my-bucket" and once in !Ref Environment
	if len(tmpl.ReferenceGraph["MyBucket"]) < 1 {
		t.Errorf("expected MyBucket to have references, got %d", len(tmpl.ReferenceGraph["MyBucket"]))
	}
}

func TestParseTemplateContent_JSON(t *testing.T) {
	tmpl, err := template.ParseTemplateContent([]byte(testJSONTemplate), "test.json")
	if err != nil {
		t.Fatalf("failed to parse JSON template: %v", err)
	}

	if tmpl.Description != "JSON test template" {
		t.Errorf("expected description, got %q", tmpl.Description)
	}

	bucket := tmpl.Resources["MyBucket"]
	if bucket == nil {
		t.Fatal("expected MyBucket resource")
	}

	bucketNameProp := bucket.Properties["BucketName"]
	if bucketNameProp == nil {
		t.Fatal("expected BucketName property")
	}
	intrinsic, ok := bucketNameProp.Value.(*template.Intrinsic)
	if !ok {
		t.Fatalf("expected BucketName to be *Intrinsic, got %T", bucketNameProp.Value)
	}
	if intrinsic.Type != template.IntrinsicSub {
		t.Errorf("expected IntrinsicSub, got %s", intrinsic.Type)
	}
}

func TestParseTemplate_File(t *testing.T) {
	tmpDir := t.TempDir()
	templatePath := filepath.Join(tmpDir, "template.yaml")
	if err := os.WriteFile(templatePath, []byte(testYAMLTemplate), 0644); err != nil {
		t.Fatalf("failed to write template file: %v", err)
	}

	tmpl, err := template.ParseTemplate(templatePath)
	if err != nil {
		t.Fatalf("failed to parse template: %v", err)
	}

	if tmpl.SourceFile != templatePath {
		t.Errorf("expected source file %q, got %q", templatePath, tmpl.SourceFile)
	}
	if len(tmpl.Resources) != 2 {
		t.Errorf("expected 2 resources, got %d", len(tmpl.Resources))
	}
}

func TestParseTemplate_NotFound(t *testing.T) {
	_, err := template.ParseTemplate("/nonexistent/template.yaml")
	if err == nil {
		t.Error("expected error for non-existent file")
	}
}

func TestParseTemplateContent_InvalidYAML(t *testing.T) {
	content := []byte("this: is: not: valid: yaml: [")
	_, err := template.ParseTemplateContent(content, "invalid.yaml")
	if err == nil {
		t.Error("expected error for invalid YAML")
	}
}

func TestParseTemplateContent_KubernetesManifest(t *testing.T) {
	content := []byte(`apiVersion: v1
kind: Pod
metadata:
  name: test
`)
	_, err := template.ParseTemplateContent(content, "k8s.yaml")
	if err == nil {
		t.Error("expected error for Kubernetes manifest")
	}
}

func TestParseTemplateContent_RainTags(t *testing.T) {
	content := []byte(`Resources:
  Bucket:
    Type: !Rain::S3 my-bucket
`)
	_, err := template.ParseTemplateContent(content, "rain.yaml")
	if err == nil {
		t.Error("expected error for Rain-specific tags")
	}
}

func TestIntrinsicType_String(t *testing.T) {
	tests := []struct {
		t    template.IntrinsicType
		want string
	}{
		{template.IntrinsicRef, "Ref"},
		{template.IntrinsicGetAtt, "GetAtt"},
		{template.IntrinsicSub, "Sub"},
		{template.IntrinsicJoin, "Join"},
		{template.IntrinsicSelect, "Select"},
		{template.IntrinsicGetAZs, "GetAZs"},
		{template.IntrinsicIf, "If"},
		{template.IntrinsicEquals, "Equals"},
		{template.IntrinsicAnd, "And"},
		{template.IntrinsicOr, "Or"},
		{template.IntrinsicNot, "Not"},
		{template.IntrinsicCondition, "Condition"},
		{template.IntrinsicFindInMap, "FindInMap"},
		{template.IntrinsicBase64, "Base64"},
		{template.IntrinsicCidr, "Cidr"},
		{template.IntrinsicImportValue, "ImportValue"},
		{template.IntrinsicSplit, "Split"},
		{template.IntrinsicTransform, "Transform"},
		{template.IntrinsicValueOf, "ValueOf"},
		{template.IntrinsicType(99), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.t.String(); got != tt.want {
				t.Errorf("IntrinsicType.String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestNewTemplate(t *testing.T) {
	tmpl := template.NewTemplate()

	if tmpl.AWSTemplateFormatVersion != "2010-09-09" {
		t.Errorf("expected default version 2010-09-09, got %s", tmpl.AWSTemplateFormatVersion)
	}
	if tmpl.Parameters == nil {
		t.Error("expected Parameters to be initialized")
	}
	if tmpl.Mappings == nil {
		t.Error("expected Mappings to be initialized")
	}
	if tmpl.Conditions == nil {
		t.Error("expected Conditions to be initialized")
	}
	if tmpl.Resources == nil {
		t.Error("expected Resources to be initialized")
	}
	if tmpl.Outputs == nil {
		t.Error("expected Outputs to be initialized")
	}
	if tmpl.ReferenceGraph == nil {
		t.Error("expected ReferenceGraph to be initialized")
	}
}

func TestResource_ServiceAndTypeName(t *testing.T) {
	tests := []struct {
		resourceType string
		wantService  string
		wantTypeName string
	}{
		{"AWS::S3::Bucket", "S3", "Bucket"},
		{"AWS::EC2::Instance", "EC2", "Instance"},
		{"AWS::Lambda::Function", "Lambda", "Function"},
		{"AWS::IAM::Role", "IAM", "Role"},
		{"InvalidType", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.resourceType, func(t *testing.T) {
			r := &template.Resource{ResourceType: tt.resourceType}
			if got := r.Service(); got != tt.wantService {
				t.Errorf("Service() = %q, want %q", got, tt.wantService)
			}
			if got := r.TypeName(); got != tt.wantTypeName {
				t.Errorf("TypeName() = %q, want %q", got, tt.wantTypeName)
			}
		})
	}
}

func TestParseTemplateContent_AllIntrinsics(t *testing.T) {
	content := []byte(`
AWSTemplateFormatVersion: "2010-09-09"
Conditions:
  Cond1: !Equals [!Ref Param, "value"]
  Cond2: !And
    - !Condition Cond1
    - !Equals [!Ref Other, "x"]
  Cond3: !Or
    - !Condition Cond1
    - !Condition Cond2
  Cond4: !Not [!Condition Cond1]
Resources:
  Test:
    Type: AWS::S3::Bucket
    Properties:
      RefProp: !Ref Param
      GetAttProp: !GetAtt Other.Arn
      GetAttDotProp: !GetAtt Other.DomainName
      SubProp: !Sub "${AWS::Region}-test"
      JoinProp: !Join ["-", ["a", "b"]]
      SelectProp: !Select [0, !GetAZs ""]
      IfProp: !If [Cond1, "yes", "no"]
      FindInMapProp: !FindInMap [Map, !Ref AWS::Region, Key]
      Base64Prop: !Base64 "data"
      CidrProp: !Cidr [!GetAtt VPC.CidrBlock, 6, 8]
      ImportProp: !ImportValue SharedStack-Value
      SplitProp: !Split [",", "a,b,c"]
`)

	tmpl, err := template.ParseTemplateContent(content, "test.yaml")
	if err != nil {
		t.Fatalf("failed to parse template: %v", err)
	}

	resource := tmpl.Resources["Test"]
	if resource == nil {
		t.Fatal("expected Test resource")
	}

	// Verify each property was parsed as an intrinsic
	props := map[string]template.IntrinsicType{
		"RefProp":       template.IntrinsicRef,
		"GetAttProp":    template.IntrinsicGetAtt,
		"GetAttDotProp": template.IntrinsicGetAtt,
		"SubProp":       template.IntrinsicSub,
		"JoinProp":      template.IntrinsicJoin,
		"SelectProp":    template.IntrinsicSelect,
		"IfProp":        template.IntrinsicIf,
		"FindInMapProp": template.IntrinsicFindInMap,
		"Base64Prop":    template.IntrinsicBase64,
		"CidrProp":      template.IntrinsicCidr,
		"ImportProp":    template.IntrinsicImportValue,
		"SplitProp":     template.IntrinsicSplit,
	}

	for propName, expectedType := range props {
		prop := resource.Properties[propName]
		if prop == nil {
			t.Errorf("expected property %s", propName)
			continue
		}
		intrinsic, ok := prop.Value.(*template.Intrinsic)
		if !ok {
			t.Errorf("expected %s to be *Intrinsic, got %T", propName, prop.Value)
			continue
		}
		if intrinsic.Type != expectedType {
			t.Errorf("expected %s type %s, got %s", propName, expectedType, intrinsic.Type)
		}
	}
}
