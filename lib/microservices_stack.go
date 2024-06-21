package lib

import (
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsapigateway"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsdynamodb"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslambda"
	"github.com/aws/aws-cdk-go/awscdk/v2/awssqs"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

type MicroservicesStackProps struct {
	awscdk.StackProps
}

func NewMicroservicesStack(scope constructs.Construct, id string, props *MicroservicesStackProps) awscdk.Stack {
	stack := awscdk.NewStack(scope, &id, &props.StackProps)

	// DynamoDB tables
	ordersTable := awsdynamodb.NewTable(stack, jsii.String("orders_table"), &awsdynamodb.TableProps{
		PartitionKey: &awsdynamodb.Attribute{
			Name: jsii.String("PK"),
			Type: awsdynamodb.AttributeType_STRING,
		},
		SortKey: &awsdynamodb.Attribute{
			Name: jsii.String("SK"),
			Type: awsdynamodb.AttributeType_STRING,
		},
		BillingMode: awsdynamodb.BillingMode_PAY_PER_REQUEST,
	})

	paymentsTable := awsdynamodb.NewTable(stack, jsii.String("payments_table"), &awsdynamodb.TableProps{
		PartitionKey: &awsdynamodb.Attribute{
			Name: jsii.String("PK"),
			Type: awsdynamodb.AttributeType_STRING,
		},
		SortKey: &awsdynamodb.Attribute{
			Name: jsii.String("SK"),
			Type: awsdynamodb.AttributeType_STRING,
		},
		BillingMode: awsdynamodb.BillingMode_PAY_PER_REQUEST,
	})

	// SQS queues
	orderCreatedQueue := awssqs.NewQueue(stack, jsii.String("OrderCreatedQueue"), nil)
	orderCompletedQueue := awssqs.NewQueue(stack, jsii.String("OrderCompletedQueue"), nil)

	// Lambda functions
	ordersHandler := awslambda.NewFunction(stack, jsii.String("OrdersHandler"), &awslambda.FunctionProps{
		Runtime: awslambda.Runtime_GO_1_X(),
		Handler: jsii.String("orders-handler"),
		Code:    awslambda.Code_FromAsset(jsii.String("../cmd/orders-handler"), nil),
		Environment: &map[string]*string{
			"DYNAMODB_TABLE_NAME": ordersTable.TableName(),
			"SQS_QUEUE_URL":       orderCreatedQueue.QueueUrl(),
		},
	})

	paymentsHandler := awslambda.NewFunction(stack, jsii.String("PaymentsHandler"), &awslambda.FunctionProps{
		Runtime: awslambda.Runtime_GO_1_X(),
		Handler: jsii.String("payments-handler"),
		Code:    awslambda.Code_FromAsset(jsii.String("../cmd/payments-handler"), nil),
		Environment: &map[string]*string{
			"DYNAMODB_TABLE_NAME": paymentsTable.TableName(),
			"SQS_QUEUE_URL":       orderCompletedQueue.QueueUrl(),
		},
	})

	// Grant permissions to Lambda functions
	ordersTable.GrantReadWriteData(ordersHandler)
	paymentsTable.GrantReadWriteData(paymentsHandler)
	orderCreatedQueue.GrantSendMessages(ordersHandler)
	orderCompletedQueue.GrantConsumeMessages(paymentsHandler)

	// API Gateway
	api := awsapigateway.NewRestApi(stack, jsii.String("OrderPaymentsAPI"), nil)

	ordersResource := api.Root().AddResource(jsii.String("orders"), nil)
	ordersResource.AddMethod(jsii.String("POST"), awsapigateway.NewLambdaIntegration(ordersHandler, nil), &awsapigateway.MethodOptions{
		AuthorizationType: awsapigateway.AuthorizationType_NONE,
	})

	paymentsResource := api.Root().AddResource(jsii.String("payments"), nil)
	paymentsResource.AddMethod(jsii.String("POST"), awsapigateway.NewLambdaIntegration(paymentsHandler, nil), &awsapigateway.MethodOptions{
		AuthorizationType: awsapigateway.AuthorizationType_NONE,
	})

	return stack
}
