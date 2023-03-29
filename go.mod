module goTest

go 1.19

require github.com/aws/aws-sdk-go v1.44.199

require github.com/jmespath/go-jmespath v0.4.0 // indirect

//require github.com/sashkoristov/fService latest
require github.com/FaaSTools/GoStorage latest

replace github.com/FaaSTools/GoStorage => ../GoStorage