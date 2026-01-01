package intrinsics_test

import (
	"testing"

	"github.com/lex00/cloudformation-schema-go/intrinsics"
)

func TestPseudoParameters(t *testing.T) {
	tests := []struct {
		name  string
		param intrinsics.Ref
		want  string
	}{
		{
			name:  "AWS_ACCOUNT_ID",
			param: intrinsics.AWS_ACCOUNT_ID,
			want:  `{"Ref":"AWS::AccountId"}`,
		},
		{
			name:  "AWS_NOTIFICATION_ARNS",
			param: intrinsics.AWS_NOTIFICATION_ARNS,
			want:  `{"Ref":"AWS::NotificationARNs"}`,
		},
		{
			name:  "AWS_NO_VALUE",
			param: intrinsics.AWS_NO_VALUE,
			want:  `{"Ref":"AWS::NoValue"}`,
		},
		{
			name:  "AWS_PARTITION",
			param: intrinsics.AWS_PARTITION,
			want:  `{"Ref":"AWS::Partition"}`,
		},
		{
			name:  "AWS_REGION",
			param: intrinsics.AWS_REGION,
			want:  `{"Ref":"AWS::Region"}`,
		},
		{
			name:  "AWS_STACK_ID",
			param: intrinsics.AWS_STACK_ID,
			want:  `{"Ref":"AWS::StackId"}`,
		},
		{
			name:  "AWS_STACK_NAME",
			param: intrinsics.AWS_STACK_NAME,
			want:  `{"Ref":"AWS::StackName"}`,
		},
		{
			name:  "AWS_URL_SUFFIX",
			param: intrinsics.AWS_URL_SUFFIX,
			want:  `{"Ref":"AWS::URLSuffix"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := mustMarshal(t, tt.param)
			if got != tt.want {
				t.Errorf("got %s, want %s", got, tt.want)
			}
		})
	}
}
