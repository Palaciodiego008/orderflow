package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

func init() {
	sess := session.Must(session.NewSession())
	dynamoDbClient = dynamodb.New(sess)
}

func HandleOrderCreatedEvent(ctx context.Context, sqsEvent events.SQSEvent) error {
	for _, message := range sqsEvent.Records {
		var orderEvent CreateOrderEvent
		err := json.Unmarshal([]byte(message.Body), &orderEvent)
		if err != nil {
			return fmt.Errorf("failed to unmarshal SQS message: %v", err)
		}

		orderItem := map[string]*dynamodb.AttributeValue{
			"OrderID": {
				S: aws.String(orderEvent.OrderID),
			},
			"OrderStatus": {
				S: aws.String("INCOMPLETE"),
			},
		}

		_, err = dynamoDbClient.PutItem(&dynamodb.PutItemInput{
			TableName: aws.String(tableName),
			Item:      orderItem,
		})
		if err != nil {
			return fmt.Errorf("failed to update order status in DynamoDB: %v", err)
		}
	}
	return nil
}
