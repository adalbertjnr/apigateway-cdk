package main

import (
	"log"
	"os"

	"github.com/aws/aws-cdk-go/awscdk/v2"

	"github.com/aws/jsii-runtime-go"
)

type CdkStackProps struct {
	awscdk.StackProps
}

const (
	FileName = "gateway_attr.yaml"

	GatewayStack = "CDKGatewayStack"
	VPCStack     = "CDKVpcLinkStack"
)

func main() {
	defer jsii.Close()
	app := awscdk.NewApp(nil)

	read := newYamlReader()
	gatewayAttrs := &ApiGatewayConfig{}

	if err := read.fromFile(FileName).
		deserializeInto(gatewayAttrs); err != nil {
		log.Fatalf("failed to deserialize gateway attribuites: %v", err)
	}

	_, vpcLink := NewVPCLinkStack(app, VPCStack, &CdkStackProps{
		awscdk.StackProps{
			Env: env(),
		}},
		gatewayAttrs.APIGatewayConfig.VpcLinkConfig,
	)

	for gatewayIndex, gatewayConfig := range gatewayAttrs.APIGatewayConfig.GatewaysConfig {
		NewApigwStack(app, GatewayStack,
			&CdkStackProps{awscdk.StackProps{Env: env()}},
			gatewayAttrs.APIGatewayConfig.Environment,
			gatewayAttrs.APIGatewayConfig.DomainConfig,
			gatewayAttrs.APIGatewayConfig.VpcLinkConfig,
			gatewayAttrs.APIGatewayConfig.IntegrationConfig,
			gatewayIndex,
			gatewayConfig,
			vpcLink,
		)
	}

	app.Synth(nil)

}

func env() *awscdk.Environment {
	return &awscdk.Environment{
		Account: jsii.String(os.Getenv("CDK_DEFAULT_ACCOUNT")),
		Region:  jsii.String(os.Getenv("CDK_DEFAULT_REGION")),
	}
}
