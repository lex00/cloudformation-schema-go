package intrinsics_test

import (
	"encoding/json"
	"testing"

	"github.com/lex00/cloudformation-schema-go/intrinsics"
)

func mustMarshal(t *testing.T, v any) string {
	t.Helper()
	data, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}
	return string(data)
}

func TestRef(t *testing.T) {
	tests := []struct {
		name string
		ref  intrinsics.Ref
		want string
	}{
		{
			name: "resource",
			ref:  intrinsics.Ref{LogicalName: "MyBucket"},
			want: `{"Ref":"MyBucket"}`,
		},
		{
			name: "parameter",
			ref:  intrinsics.Param("VpcId"),
			want: `{"Ref":"VpcId"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := mustMarshal(t, tt.ref)
			if got != tt.want {
				t.Errorf("got %s, want %s", got, tt.want)
			}
		})
	}
}

func TestGetAtt(t *testing.T) {
	g := intrinsics.GetAtt{LogicalName: "MyBucket", Attribute: "Arn"}
	got := mustMarshal(t, g)
	want := `{"Fn::GetAtt":["MyBucket","Arn"]}`
	if got != want {
		t.Errorf("got %s, want %s", got, want)
	}
}

func TestSub(t *testing.T) {
	s := intrinsics.Sub{String: "${AWS::Region}-bucket"}
	got := mustMarshal(t, s)
	want := `{"Fn::Sub":"${AWS::Region}-bucket"}`
	if got != want {
		t.Errorf("got %s, want %s", got, want)
	}
}

func TestSubWithMap(t *testing.T) {
	s := intrinsics.SubWithMap{
		String: "${Bucket}-data",
		Variables: map[string]any{
			"Bucket": intrinsics.Ref{LogicalName: "MyBucket"},
		},
	}
	got := mustMarshal(t, s)
	// Note: map ordering in JSON may vary, so we unmarshal and check structure
	var result map[string][]any
	if err := json.Unmarshal([]byte(got), &result); err != nil {
		t.Fatalf("failed to unmarshal result: %v", err)
	}
	if len(result["Fn::Sub"]) != 2 {
		t.Errorf("expected 2 elements in Fn::Sub array, got %d", len(result["Fn::Sub"]))
	}
}

func TestJoin(t *testing.T) {
	j := intrinsics.Join{Delimiter: ",", Values: []any{"a", "b", "c"}}
	got := mustMarshal(t, j)
	want := `{"Fn::Join":[",",["a","b","c"]]}`
	if got != want {
		t.Errorf("got %s, want %s", got, want)
	}
}

func TestSelect(t *testing.T) {
	s := intrinsics.Select{Index: 0, List: intrinsics.GetAZs{Region: ""}}
	got := mustMarshal(t, s)
	want := `{"Fn::Select":[0,{"Fn::GetAZs":""}]}`
	if got != want {
		t.Errorf("got %s, want %s", got, want)
	}
}

func TestGetAZs(t *testing.T) {
	tests := []struct {
		name   string
		getAZs intrinsics.GetAZs
		want   string
	}{
		{
			name:   "current_region",
			getAZs: intrinsics.GetAZs{Region: ""},
			want:   `{"Fn::GetAZs":""}`,
		},
		{
			name:   "specific_region",
			getAZs: intrinsics.GetAZs{Region: "us-east-1"},
			want:   `{"Fn::GetAZs":"us-east-1"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := mustMarshal(t, tt.getAZs)
			if got != tt.want {
				t.Errorf("got %s, want %s", got, tt.want)
			}
		})
	}
}

func TestIf(t *testing.T) {
	i := intrinsics.If{
		Condition:    "CreateResources",
		ValueIfTrue:  "yes",
		ValueIfFalse: "no",
	}
	got := mustMarshal(t, i)
	want := `{"Fn::If":["CreateResources","yes","no"]}`
	if got != want {
		t.Errorf("got %s, want %s", got, want)
	}
}

func TestEquals(t *testing.T) {
	e := intrinsics.Equals{Value1: intrinsics.Ref{LogicalName: "Env"}, Value2: "prod"}
	got := mustMarshal(t, e)
	want := `{"Fn::Equals":[{"Ref":"Env"},"prod"]}`
	if got != want {
		t.Errorf("got %s, want %s", got, want)
	}
}

func TestAnd(t *testing.T) {
	a := intrinsics.And{Conditions: []any{
		intrinsics.Condition{Name: "Cond1"},
		intrinsics.Condition{Name: "Cond2"},
	}}
	got := mustMarshal(t, a)
	want := `{"Fn::And":[{"Condition":"Cond1"},{"Condition":"Cond2"}]}`
	if got != want {
		t.Errorf("got %s, want %s", got, want)
	}
}

func TestOr(t *testing.T) {
	o := intrinsics.Or{Conditions: []any{
		intrinsics.Condition{Name: "Cond1"},
		intrinsics.Condition{Name: "Cond2"},
	}}
	got := mustMarshal(t, o)
	want := `{"Fn::Or":[{"Condition":"Cond1"},{"Condition":"Cond2"}]}`
	if got != want {
		t.Errorf("got %s, want %s", got, want)
	}
}

func TestNot(t *testing.T) {
	n := intrinsics.Not{Condition: intrinsics.Condition{Name: "Cond1"}}
	got := mustMarshal(t, n)
	want := `{"Fn::Not":[{"Condition":"Cond1"}]}`
	if got != want {
		t.Errorf("got %s, want %s", got, want)
	}
}

func TestBase64(t *testing.T) {
	b := intrinsics.Base64{Value: "Hello World"}
	got := mustMarshal(t, b)
	want := `{"Fn::Base64":"Hello World"}`
	if got != want {
		t.Errorf("got %s, want %s", got, want)
	}
}

func TestImportValue(t *testing.T) {
	i := intrinsics.ImportValue{ExportName: "SharedVpcId"}
	got := mustMarshal(t, i)
	want := `{"Fn::ImportValue":"SharedVpcId"}`
	if got != want {
		t.Errorf("got %s, want %s", got, want)
	}
}

func TestFindInMap(t *testing.T) {
	f := intrinsics.FindInMap{
		MapName:   "RegionMap",
		TopKey:    intrinsics.Ref{LogicalName: "AWS::Region"},
		SecondKey: "AMI",
	}
	got := mustMarshal(t, f)
	want := `{"Fn::FindInMap":["RegionMap",{"Ref":"AWS::Region"},"AMI"]}`
	if got != want {
		t.Errorf("got %s, want %s", got, want)
	}
}

func TestSplit(t *testing.T) {
	s := intrinsics.Split{Delimiter: ",", Source: "a,b,c"}
	got := mustMarshal(t, s)
	want := `{"Fn::Split":[",","a,b,c"]}`
	if got != want {
		t.Errorf("got %s, want %s", got, want)
	}
}

func TestCidr(t *testing.T) {
	c := intrinsics.Cidr{
		IPBlock:  intrinsics.GetAtt{LogicalName: "VPC", Attribute: "CidrBlock"},
		Count:    6,
		CidrBits: "8",
	}
	got := mustMarshal(t, c)
	want := `{"Fn::Cidr":[{"Fn::GetAtt":["VPC","CidrBlock"]},6,"8"]}`
	if got != want {
		t.Errorf("got %s, want %s", got, want)
	}
}

func TestCondition(t *testing.T) {
	c := intrinsics.Condition{Name: "CreateResources"}
	got := mustMarshal(t, c)
	want := `{"Condition":"CreateResources"}`
	if got != want {
		t.Errorf("got %s, want %s", got, want)
	}
}

func TestTag(t *testing.T) {
	tests := []struct {
		name string
		tag  intrinsics.Tag
		want string
	}{
		{
			name: "string_value",
			tag:  intrinsics.Tag{Key: "Name", Value: "my-resource"},
			want: `{"Key":"Name","Value":"my-resource"}`,
		},
		{
			name: "intrinsic_value",
			tag:  intrinsics.Tag{Key: "Environment", Value: intrinsics.Ref{LogicalName: "Env"}},
			want: `{"Key":"Environment","Value":{"Ref":"Env"}}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := mustMarshal(t, tt.tag)
			if got != tt.want {
				t.Errorf("got %s, want %s", got, tt.want)
			}
		})
	}
}

func TestTransform(t *testing.T) {
	tr := intrinsics.Transform{
		Name: "AWS::Include",
		Parameters: map[string]any{
			"Location": "s3://bucket/template.yaml",
		},
	}
	got := mustMarshal(t, tr)
	// Map ordering may vary, so we just check it contains the right structure
	var result map[string]any
	if err := json.Unmarshal([]byte(got), &result); err != nil {
		t.Fatalf("failed to unmarshal result: %v", err)
	}
	transform, ok := result["Fn::Transform"].(map[string]any)
	if !ok {
		t.Fatal("expected Fn::Transform to be a map")
	}
	if transform["Name"] != "AWS::Include" {
		t.Errorf("expected Name to be AWS::Include, got %v", transform["Name"])
	}
}

func TestOutput(t *testing.T) {
	tests := []struct {
		name   string
		output intrinsics.Output
		check  func(t *testing.T, result map[string]any)
	}{
		{
			name: "minimal",
			output: intrinsics.Output{
				Value: intrinsics.Ref{LogicalName: "MyBucket"},
			},
			check: func(t *testing.T, result map[string]any) {
				if result["Value"] == nil {
					t.Error("expected Value to be set")
				}
				if result["Description"] != nil {
					t.Error("expected Description to be nil")
				}
			},
		},
		{
			name: "with_description",
			output: intrinsics.Output{
				Value:       intrinsics.GetAtt{LogicalName: "MyBucket", Attribute: "Arn"},
				Description: "The ARN of the bucket",
			},
			check: func(t *testing.T, result map[string]any) {
				if result["Description"] != "The ARN of the bucket" {
					t.Errorf("expected Description to be set, got %v", result["Description"])
				}
			},
		},
		{
			name: "with_export",
			output: intrinsics.Output{
				Value:      intrinsics.Ref{LogicalName: "MyBucket"},
				ExportName: intrinsics.Sub{String: "${AWS::StackName}-BucketName"},
			},
			check: func(t *testing.T, result map[string]any) {
				export, ok := result["Export"].(map[string]any)
				if !ok {
					t.Fatal("expected Export to be a map")
				}
				if export["Name"] == nil {
					t.Error("expected Export.Name to be set")
				}
			},
		},
		{
			name: "with_condition",
			output: intrinsics.Output{
				Value:     intrinsics.Ref{LogicalName: "MyBucket"},
				Condition: "CreateBucket",
			},
			check: func(t *testing.T, result map[string]any) {
				if result["Condition"] != "CreateBucket" {
					t.Errorf("expected Condition to be CreateBucket, got %v", result["Condition"])
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := mustMarshal(t, tt.output)
			var result map[string]any
			if err := json.Unmarshal([]byte(data), &result); err != nil {
				t.Fatalf("failed to unmarshal result: %v", err)
			}
			tt.check(t, result)
		})
	}
}
