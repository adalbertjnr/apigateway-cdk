package main

import (
	"fmt"
	"strings"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsapigatewayv2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awscertificatemanager"
	"github.com/aws/aws-cdk-go/awscdk/v2/awss3"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

type Mtls struct {
	Bucket  string `yaml:"bucket"`
	Key     string `yaml:"key"`
	Version string `yaml:"version"`
}

type Logging struct {
	LoggerArn string `yaml:"loggerArn"`
}

type ApiGatewayConfig struct {
	APIGatewayConfig ApiGatewayAttributes `yaml:"apiGatewayConfig"`
}

type ApiGatewayAttributes struct {
	Environment string `yaml:"environment"`

	DomainConfig      DomainConfig      `yaml:"domainConfig"`
	VpcLinkConfig     VpcLinkConfig     `yaml:"vpcLinkConfig"`
	IntegrationConfig IntegrationConfig `yaml:"integrationConfig"`
	GatewaysConfig    []GatewayConfig   `yaml:"gatewaysConfig"`
}

type DomainConfig struct {
	Name         string `yaml:"name"`
	AcmArn       string `yaml:"acmArn"`
	HostedZoneId string `yaml:"hostedZoneId"`
}

type VpcLinkConfig struct {
	Name           string   `yaml:"name"`
	VpcId          string   `yaml:"vpcId"`
	VpcLinkId      *string  `yaml:"vpcLinkId,omitempty"`
	Subnets        []string `yaml:"subnets"`
	SecurityGroups []string `yaml:"securityGroups"`
}

type IntegrationConfig struct {
	AlbListenerArn string `yaml:"albListenerArn"`
}

type GatewayConfig struct {
	AppName string   `yaml:"appName"`
	Mtls    *Mtls    `yaml:"mtls,omitempty"`
	Logging *Logging `yaml:"logging,omitempty"`
	Routes  []Routes `yaml:"routes"`
}

type Routes struct {
	Path        string   `yaml:"path"`
	Methods     []string `yaml:"methods"`
	Integration *string  `yaml:"integration,omitempty"`
}

type stackManager struct {
	name       awscdk.Stack
	scope      constructs.Construct
	domainName string
}

func newManagedStack(scope constructs.Construct, id *string, props *awscdk.StackProps) *stackManager {
	stack := awscdk.NewStack(scope, id, props)
	return &stackManager{
		name:  stack,
		scope: scope,
	}
}

func (m *stackManager) createDomainFrom(appName, domain string) *stackManager {
	m.domainName = fmt.Sprintf("%s.%s", strings.ToLower(appName), domain)
	return m
}

func (m *stackManager) withCert(cert awscertificatemanager.ICertificate, tls *Mtls, id string) awsapigatewayv2.DomainName {
	var tlsConfig *awsapigatewayv2.MTLSConfig
	if tls != nil {
		tlsConfig = &awsapigatewayv2.MTLSConfig{
			Bucket:  awss3.Bucket_FromBucketName(m.name, jsii.String(id+"-bucket"), &tls.Bucket),
			Key:     &tls.Key,
			Version: &tls.Version,
		}
	}
	return awsapigatewayv2.NewDomainName(m.name, jsii.String(id+"-domain"), &awsapigatewayv2.DomainNameProps{
		Certificate: cert,
		DomainName:  &m.domainName,
		Mtls:        tlsConfig,
	})
}
