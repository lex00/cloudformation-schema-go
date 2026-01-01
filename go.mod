module github.com/lex00/cloudformation-schema-go

go 1.23

require (
	github.com/aws/aws-sdk-go-v2/service/dynamodb v1.53.5
	github.com/aws/aws-sdk-go-v2/service/ec2 v1.279.0
	github.com/aws/aws-sdk-go-v2/service/ecs v1.70.0
	github.com/aws/aws-sdk-go-v2/service/lambda v1.87.0
	github.com/aws/aws-sdk-go-v2/service/s3 v1.95.0
	gopkg.in/yaml.v3 v3.0.1
)

require github.com/aws/smithy-go v1.24.0 // indirect
