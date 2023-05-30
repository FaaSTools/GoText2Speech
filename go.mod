module goTest

go 1.19

require (
	github.com/aws/aws-sdk-go v1.44.199
	github.com/aws/aws-sdk-go-v2 v1.18.0
)

require (
	github.com/aws/smithy-go v1.13.5 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
)

replace github.com/FaaSTools/GoStorage => ../GoStorage
