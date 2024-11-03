package main

import (
	"fmt"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsapigatewayv2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsec2"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

func NewVPCLinkStack(scope constructs.Construct, id string, props *CdkStackProps, vpcLinkConfig VpcLinkConfig) (awscdk.Stack, awsapigatewayv2.IVpcLink) {
	var sprops awscdk.StackProps
	if props != nil {
		sprops = props.StackProps
	}

	stack := newManagedStack(scope, &id, &sprops)

	if !vpcLinkConfig.valid() {
		panic("Empty SecurityGroups or Subnets. You must to set at least one of them")
	}

	var vpcId = awsec2.Vpc_FromLookup(stack.name, jsii.String(id+"-vpc"), &awsec2.VpcLookupOptions{
		VpcId: &vpcLinkConfig.VpcId,
	})

	if vpcLinkConfig.VpcLinkId != nil {
		return stack.name, awsapigatewayv2.VpcLink_FromVpcLinkAttributes(
			stack.name,
			jsii.String(id+"ImportVpcLink"),
			&awsapigatewayv2.VpcLinkAttributes{
				Vpc:       vpcId,
				VpcLinkId: vpcLinkConfig.VpcLinkId,
			},
		)
	}

	subnets := vpcLinkConfig.extractSubnets(stack, &id)
	securityGroups := vpcLinkConfig.extractSecurityGroups(stack, &id)

	return stack.name, awsapigatewayv2.NewVpcLink(stack.name, jsii.String(id+"-VpcLink"), &awsapigatewayv2.VpcLinkProps{
		VpcLinkName:    &vpcLinkConfig.Name,
		Vpc:            vpcId,
		SecurityGroups: &securityGroups,
		Subnets: &awsec2.SubnetSelection{
			Subnets: &subnets,
		},
	})
}

func (vlc *VpcLinkConfig) extractSecurityGroups(stack *stackManager, id *string) []awsec2.ISecurityGroup {
	securityGroupList := make([]awsec2.ISecurityGroup, len(vlc.SecurityGroups))

	for i, securityGroupID := range vlc.SecurityGroups {
		securityGroupUniqueID := fmt.Sprintf("SecurityGroup-%s-%s-%d", *id, stack.name, i)
		securityGroupList[i] = awsec2.SecurityGroup_FromSecurityGroupId(
			stack.name,
			jsii.String(securityGroupUniqueID),
			&securityGroupID,
			&awsec2.SecurityGroupImportOptions{},
		)
	}

	return securityGroupList
}

func (vlc *VpcLinkConfig) extractSubnets(stack *stackManager, id *string) []awsec2.ISubnet {
	subnetList := make([]awsec2.ISubnet, len(vlc.Subnets))

	for i, subnetID := range vlc.Subnets {
		subnetUniqueID := fmt.Sprintf("Subnet-%s-%s-%d", *id, stack.name, i)
		subnet := awsec2.Subnet_FromSubnetAttributes(stack.name, jsii.String(subnetUniqueID), &awsec2.SubnetAttributes{
			SubnetId: &subnetID,
		})
		awscdk.Annotations_Of(subnet).AcknowledgeWarning(jsii.String("@aws-cdk/aws-ec2:noSubnetRouteTableId"), nil)
		subnetList[i] = subnet
	}

	return subnetList
}

func (vlc *VpcLinkConfig) valid() bool {
	return len(vlc.SecurityGroups) != 0 && len(vlc.Subnets) != 0
}
