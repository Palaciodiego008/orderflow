package main

import (
	"microservices-infra/lib"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/jsii-runtime-go"
)

func main() {
	app := awscdk.NewApp(nil)

	lib.NewMicroservicesStack(app, "MicroservicesStack", &lib.MicroservicesStackProps{
		StackProps: awscdk.StackProps{
			Env: env(),
		},
	})

	app.Synth(nil)
}

func env() *awscdk.Environment {
	return &awscdk.Environment{
		Account: jsii.String("129260641130"),
		Region:  jsii.String("us-east-1"),
	}
}
