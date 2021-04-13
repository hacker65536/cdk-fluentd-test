package main

import (
	"os"

	"github.com/aws/aws-cdk-go/awscdk"
	"github.com/aws/aws-cdk-go/awscdk/awsec2"
	"github.com/aws/aws-cdk-go/awscdk/awsiam"
	"github.com/aws/constructs-go/constructs/v3"
	"github.com/aws/jsii-runtime-go"
)

type CdkFluentdTestStackProps struct {
	awscdk.StackProps
}

func NewCdkFluentdTestStack(scope constructs.Construct, id string, props *CdkFluentdTestStackProps) awscdk.Stack {
	var sprops awscdk.StackProps
	if props != nil {
		sprops = props.StackProps
	}
	stack := awscdk.NewStack(scope, &id, &sprops)

	// The code that defines your stack goes here

	// VPC
	vpc := awsec2.NewVpc(stack, jsii.String("myvpc"), &awsec2.VpcProps{
		MaxAzs:      jsii.Number(1),
		NatGateways: jsii.Number(0),
	})

	// user_data
	//userdatacmd := "yum install -y fluentd"
	userdatacmds := []string{
		"curl -L https://toolbelt.treasuredata.com/sh/install-amazon2-td-agent4.sh | sh",
		"amazon-linux-extras install ruby2.6",
		"yum install -y ruby-devel gcc",
		"systemctl start td-agent",
		"sudo -u ec2-user -i mkdir rubyapp",
		"sudo gem install bundler",
		`sed -r -e 's|^(Defaults\s+secure_path.*)|\1:/usr/local/bin|' -i /etc/sudoers`,
		`sudo -u ec2-user -i bash -c "echo \"source 'https://rubygems.org'\">rubyapp/Gemfile"`,
		`sudo -u ec2-user -i bash -c "echo gem \"'fluent-logger', \\\"~> 0.7.1\\\"\">>rubyapp/Gemfile"`,
		`sudo -u ec2-user -i bash -c "cd rubyapp; bundle install --path vendor/bundle"`,
	}
	userdata := awsec2.UserData_ForLinux(&awsec2.LinuxUserDataOptions{
		Shebang: jsii.String("#!/usr/bin/env bash"),
	})
	for _, v := range userdatacmds {
		userdata.AddCommands(&v)
	}

	machineImage := awsec2.MachineImage_LatestAmazonLinux(&awsec2.AmazonLinuxImageProps{
		Generation: awsec2.AmazonLinuxGeneration_AMAZON_LINUX_2,
		Storage:    awsec2.AmazonLinuxStorage_GENERAL_PURPOSE,
	})
	instanceType := awsec2.InstanceType_Of(
		awsec2.InstanceClass_COMPUTE5,
		awsec2.InstanceSize_XLARGE,
	)

	// iam
	policies := []string{"AmazonSSMManagedInstanceCore"}

	// ec2 instance
	fec2 := awsec2.NewInstance(stack, jsii.String("fluenttest"), &awsec2.InstanceProps{
		Vpc: vpc,
		VpcSubnets: &awsec2.SubnetSelection{
			SubnetType: awsec2.SubnetType_PUBLIC,
		},
		UserData:     userdata,
		MachineImage: machineImage,
		InstanceType: instanceType,
	})

	for _, v := range policies {
		fec2.Role().AddManagedPolicy(
			awsiam.ManagedPolicy_FromAwsManagedPolicyName(&v),
		)
	}

	instanceid := fec2.InstanceId()
	sessioncmd := "aws ssm start-session --target "
	sessioncmd += *instanceid
	awscdk.NewCfnOutput(stack, jsii.String("fec2"), &awscdk.CfnOutputProps{
		Value: jsii.String(sessioncmd),
	})

	return stack
}

func main() {
	app := awscdk.NewApp(nil)

	NewCdkFluentdTestStack(app, "CdkFluentdTestStack", &CdkFluentdTestStackProps{
		awscdk.StackProps{
			Env: env(),
		},
	})

	app.Synth(nil)
}

// env determines the AWS environment (account+region) in which our stack is to
// be deployed. For more information see: https://docs.aws.amazon.com/cdk/latest/guide/environments.html
func env() *awscdk.Environment {
	// If unspecified, this stack will be "environment-agnostic".
	// Account/Region-dependent features and context lookups will not work, but a
	// single synthesized template can be deployed anywhere.
	//---------------------------------------------------------------------------
	//return nil

	// Uncomment if you know exactly what account and region you want to deploy
	// the stack to. This is the recommendation for production stacks.
	//---------------------------------------------------------------------------
	// return &awscdk.Environment{
	//  Account: jsii.String("123456789012"),
	//  Region:  jsii.String("us-east-1"),
	// }

	// Uncomment to specialize this stack for the AWS Account and Region that are
	// implied by the current CLI configuration. This is recommended for dev
	// stacks.
	//---------------------------------------------------------------------------
	return &awscdk.Environment{
		Account: jsii.String(os.Getenv("CDK_DEFAULT_ACCOUNT")),
		Region:  jsii.String(os.Getenv("CDK_DEFAULT_REGION")),
	}
}
