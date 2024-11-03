package main

import (
	"fmt"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsapigatewayv2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awscertificatemanager"
	"github.com/aws/aws-cdk-go/awscdk/v2/awselasticloadbalancingv2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsroute53"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsroute53targets"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

func configureRouteWithIntegration(stack *stackManager, listener awselasticloadbalancingv2.IApplicationListener, gateway awsapigatewayv2.CfnApi, route Routes, routeIndex int, baseID string, vpcLink awsapigatewayv2.IVpcLink) {
	for methodIndex, method := range route.Methods {
		uniqueID := fmt.Sprintf("%s-%d-%s-%d", baseID, methodIndex, method, routeIndex)

		integrationMapping := createParameterMapping(route.Integration)
		cfnIntegration := awsapigatewayv2.NewCfnIntegration(
			stack.name,
			jsii.String(uniqueID+"-integration"),
			&awsapigatewayv2.CfnIntegrationProps{
				ApiId:                gateway.AttrApiId(),
				IntegrationType:      jsii.String("HTTP_PROXY"),
				ConnectionId:         vpcLink.VpcLinkId(),
				ConnectionType:       jsii.String("VPC_LINK"),
				IntegrationMethod:    &method,
				IntegrationUri:       listener.ListenerArn(),
				RequestParameters:    integrationMapping,
				PayloadFormatVersion: jsii.String("1.0"),
			},
		)

		routeKey := fmt.Sprintf("%s %s", method, route.Path)
		integrationKey := fmt.Sprintf("integrations/%s", *cfnIntegration.Ref())

		awsapigatewayv2.NewCfnRoute(stack.name, jsii.String(uniqueID+"-Route"), &awsapigatewayv2.CfnRouteProps{
			ApiId:    gateway.AttrApiId(),
			RouteKey: &routeKey,
			Target:   &integrationKey,
		})

	}
}

func registerDomainNames(stack *stackManager, domain awsapigatewayv2.DomainName, domainName string, gattrValues GatewayConfig) {
	hostedZone := awsroute53.HostedZone_FromLookup(stack.name, jsii.String(gattrValues.AppName+"-HostedZone"), &awsroute53.HostedZoneProviderProps{
		DomainName: jsii.String(domainName),
	})

	awsroute53.NewARecord(stack.name, jsii.String(gattrValues.AppName+"-R53Record"), &awsroute53.ARecordProps{
		Zone:       hostedZone,
		Comment:    jsii.String("DNS Record created for " + *domain.Name()),
		RecordName: jsii.String(gattrValues.AppName),
		Target:     awsroute53.RecordTarget_FromAlias(awsroute53targets.NewApiGatewayv2DomainProperties(domain.RegionalDomainName(), domain.RegionalHostedZoneId())),
	})
}

func createHTTPAPIGateway(stack *stackManager, baseID string, domain awsapigatewayv2.DomainName, gatewayConfig GatewayConfig) awsapigatewayv2.CfnApi {
	gatewayCfn := awsapigatewayv2.NewCfnApi(stack.name, jsii.String(baseID+"-Gateway"), &awsapigatewayv2.CfnApiProps{
		Name:                      jsii.String(gatewayConfig.AppName),
		DisableExecuteApiEndpoint: jsii.Bool(true),
		ProtocolType:              jsii.String("HTTP"),
		Description:               jsii.String("Created by aws cdk for " + gatewayConfig.AppName),
	})

	var loggerConfigSettings *awsapigatewayv2.CfnStage_AccessLogSettingsProperty
	if gatewayConfig.Logging != nil && gatewayConfig.Logging.LoggerArn != "" {
		loggerConfigSettings = &awsapigatewayv2.CfnStage_AccessLogSettingsProperty{
			DestinationArn: &gatewayConfig.Logging.LoggerArn,
			Format:         jsii.String(`{"requestId": "$context.requestId", "ip": "$context.identity.sourceIp", "caller": "$context.identity.caller", "user": "$context.identity.user", "requestTime": "$context.requestTime", "httpMethod": "$context.httpMethod", "resourcePath": "$context.resourcePath", "status": "$context.status", "protocol": "$context.protocol", "responseLength": "$context.responseLength"}`),
		}
	}

	stageCfn := awsapigatewayv2.NewCfnStage(stack.name, jsii.String(baseID+"-Stage"), &awsapigatewayv2.CfnStageProps{
		ApiId:             gatewayCfn.AttrApiId(),
		AutoDeploy:        jsii.Bool(true),
		Description:       jsii.String("Stage for " + gatewayConfig.AppName),
		StageName:         jsii.String(gatewayConfig.AppName + "HTTPStage"),
		AccessLogSettings: loggerConfigSettings,
	})

	awsapigatewayv2.NewCfnApiMapping(stack.name, jsii.String(baseID+"-ApiMapping"), &awsapigatewayv2.CfnApiMappingProps{
		ApiId:      gatewayCfn.AttrApiId(),
		Stage:      stageCfn.Ref(),
		DomainName: jsii.String(*domain.Name()),
	})

	return gatewayCfn
}

func NewApigwStack(scope constructs.Construct, id string, props *CdkStackProps, environment string, domainConfig DomainConfig, vpcLinkConfig VpcLinkConfig, integrationConfig IntegrationConfig, gatewayIndex string, gatewaysAttr GatewayConfig, vpcLink awsapigatewayv2.IVpcLink) awscdk.Stack {
	var sprops awscdk.StackProps
	if props != nil {
		sprops = props.StackProps
	}

	apiStack := newManagedStack(scope, jsii.String(fmt.Sprintf("%s-%s", id, gatewayIndex)), &sprops)

	gatewayBaseID := fmt.Sprintf("%s-%s", gatewayIndex, id)

	httpAPIGatewayDomain := apiStack.createDomainFrom(gatewaysAttr.AppName, domainConfig.Name).
		withCert(
			awscertificatemanager.Certificate_FromCertificateArn(apiStack.name, jsii.String(gatewayBaseID+"-Cert"), jsii.String(domainConfig.AcmArn)),
			gatewaysAttr.Mtls,
			gatewayBaseID,
		)

	httpAPIGateway := createHTTPAPIGateway(apiStack, gatewayBaseID, httpAPIGatewayDomain, gatewaysAttr)
	registerDomainNames(apiStack, httpAPIGatewayDomain, domainConfig.Name, gatewaysAttr)

	listenerLookupOptions := awselasticloadbalancingv2.ApplicationListenerLookupOptions{ListenerArn: jsii.String(integrationConfig.AlbListenerArn)}
	listener := awselasticloadbalancingv2.ApplicationListener_FromLookup(apiStack.name, jsii.String(gatewayBaseID+"-ListenerLookup"), &listenerLookupOptions)

	for routeIndex, route := range gatewaysAttr.Routes {
		configureRouteWithIntegration(
			apiStack,
			listener,
			httpAPIGateway,
			route,
			routeIndex,
			gatewayBaseID,
			vpcLink,
		)
	}

	return apiStack.name
}

func createParameterMapping(integration *string) map[string]interface{} {
	if integration != nil {
		return map[string]interface{}{
			"overwrite:path": &integration,
		}
	}
	return map[string]interface{}{
		"overwrite:path": "",
	}
}
