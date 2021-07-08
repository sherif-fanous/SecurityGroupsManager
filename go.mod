module security-groups-manager

require (
	github.com/aws/aws-lambda-go v1.23.0
	github.com/aws/aws-sdk-go-v2 v1.6.0
	github.com/aws/aws-sdk-go-v2/config v1.3.0
	github.com/aws/aws-sdk-go-v2/service/ec2 v1.9.0
	github.com/go-test/deep v1.0.7
	github.com/jedib0t/go-pretty/v6 v6.2.2
	github.com/stretchr/testify v1.6.1
	inet.af/netaddr v0.0.0-20210603230628-bf05d8b52dda
)

go 1.16
