package main

import (
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsiam"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslambda"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslambdaeventsources"
	"github.com/aws/aws-cdk-go/awscdk/v2/awss3"
	"github.com/aws/aws-cdk-go/awscdklambdagoalpha/v2"

	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

const functionDir = "../function"

type LambdaPollyTextToSpeechGolangStackProps struct {
	awscdk.StackProps
}

func NewLambdaPollyTextToSpeechGolangStack(scope constructs.Construct, id string, props *LambdaPollyTextToSpeechGolangStackProps) awscdk.Stack {
	var sprops awscdk.StackProps
	if props != nil {
		sprops = props.StackProps
	}
	stack := awscdk.NewStack(scope, &id, &sprops)

	sourceBucket := awss3.NewBucket(stack, jsii.String("source-bucket"), &awss3.BucketProps{
		BlockPublicAccess: awss3.BlockPublicAccess_BLOCK_ALL(),
		RemovalPolicy:     awscdk.RemovalPolicy_DESTROY,
		AutoDeleteObjects: jsii.Bool(true),
	})

	targetBucket := awss3.NewBucket(stack, jsii.String("target-bucket"), &awss3.BucketProps{
		BlockPublicAccess: awss3.BlockPublicAccess_BLOCK_ALL(),
		RemovalPolicy:     awscdk.RemovalPolicy_DESTROY,
		AutoDeleteObjects: jsii.Bool(true),
	})

	function := awscdklambdagoalpha.NewGoFunction(stack, jsii.String("text-to-speech-function"),
		&awscdklambdagoalpha.GoFunctionProps{
			Runtime:     awslambda.Runtime_GO_1_X(),
			Environment: &map[string]*string{"TARGET_BUCKET_NAME": targetBucket.BucketName()},
			Entry:       jsii.String(functionDir),
		})

	//warning: the roles that are granted are more than what's required.
	sourceBucket.GrantRead(function, "*")
	targetBucket.GrantWrite(function, "*")
	function.Role().AddManagedPolicy(awsiam.ManagedPolicy_FromAwsManagedPolicyName(jsii.String("AmazonPollyReadOnlyAccess")))

	function.AddEventSource(awslambdaeventsources.NewS3EventSource(sourceBucket, &awslambdaeventsources.S3EventSourceProps{
		Events: &[]awss3.EventType{awss3.EventType_OBJECT_CREATED},
	}))

	awscdk.NewCfnOutput(stack, jsii.String("source-bucket-name"),
		&awscdk.CfnOutputProps{
			ExportName: jsii.String("source-bucket-name"),
			Value:      sourceBucket.BucketName()})

	awscdk.NewCfnOutput(stack, jsii.String("target-bucket-name"),
		&awscdk.CfnOutputProps{
			ExportName: jsii.String("target-bucket-name"),
			Value:      targetBucket.BucketName()})

	return stack
}

func main() {
	app := awscdk.NewApp(nil)

	NewLambdaPollyTextToSpeechGolangStack(app, "LambdaPollyTextToSpeechGolangStack", &LambdaPollyTextToSpeechGolangStackProps{
		awscdk.StackProps{
			Env: env(),
		},
	})

	app.Synth(nil)
}

func env() *awscdk.Environment {
	return nil
}
