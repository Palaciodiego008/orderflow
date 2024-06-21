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
		Account: jsii.String("YOUR_ACCOUNT_ID"),
		Region:  jsii.String("YOUR_REGION"),
	}
}
