package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/sqs"
)

type ProcessPaymentRequest struct {
	OrderID string `json:"order_id"`
	Status  string `json:"status"`
}

type OrderCompletedEvent struct {
	OrderID string `json:"order_id"`
}

var (
	dynamoDBTableName string
	sqsQueueURL       string
)

func init() {
	// Initialize DynamoDB table name and SQS queue URL from environment variables
	dynamoDBTableName = os.Getenv("DYNAMODB_TABLE_NAME")
	sqsQueueURL = os.Getenv("SQS_QUEUE_URL")

	if dynamoDBTableName == "" {
		log.Fatal("DYNAMODB_TABLE_NAME environment variable not set")
	}

	if sqsQueueURL == "" {
		log.Fatal("SQS_QUEUE_URL environment variable not set")
	}
}

func processPaymentHandler(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// Parse the request body
	var req ProcessPaymentRequest
	err := json.Unmarshal([]byte(request.Body), &req)
	if err != nil {
		return events.APIGatewayProxyResponse{StatusCode: 400, Body: "Invalid request body"}, nil
	}

	log.Printf("Processing payment for order ID: %s, Status: %s", req.OrderID, req.Status)

	// Update payment status in DynamoDB
	err = updatePaymentStatus(req.OrderID, req.Status)
	if err != nil {
		return events.APIGatewayProxyResponse{StatusCode: 500, Body: "Failed to update payment status"}, nil
	}

	// Send order completed event to Orders service via SQS
	err = sendOrderCompletedEvent(req.OrderID)
	if err != nil {
		return events.APIGatewayProxyResponse{StatusCode: 500, Body: "Failed to send order completed event"}, nil
	}

	// Return success response
	responseBody := fmt.Sprintf("Payment processed successfully for order ID: %s", req.OrderID)
	return events.APIGatewayProxyResponse{StatusCode: 200, Body: responseBody}, nil
}

func updatePaymentStatus(orderID string, status string) error {
	// Create a new AWS session and DynamoDB client
	sess := session.Must(session.NewSession())
	svc := dynamodb.New(sess)

	// UpdateItem input
	input := &dynamodb.UpdateItemInput{
		TableName: aws.String(dynamoDBTableName),
		Key: map[string]*dynamodb.AttributeValue{
			"PK": {S: aws.String(fmt.Sprintf("ORDER#%s", orderID))},
			"SK": {S: aws.String("DETAILS")},
		},
		UpdateExpression: aws.String("SET payment_status = :status"),
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":status": {S: aws.String(status)},
		},
	}

	// Update item in DynamoDB
	_, err := svc.UpdateItem(input)
	if err != nil {
		return err
	}

	return nil
}

func sendOrderCompletedEvent(orderID string) error {
	// Create a new AWS session and SQS client
	sess := session.Must(session.NewSession())
	sqsSvc := sqs.New(sess)

	// Prepare event data
	event := OrderCompletedEvent{
		OrderID: orderID,
	}
	eventJSON, err := json.Marshal(event)
	if err != nil {
		return err
	}

	// Send message to SQS queue
	_, err = sqsSvc.SendMessage(&sqs.SendMessageInput{
		MessageBody: aws.String(string(eventJSON)),
		QueueUrl:    aws.String(sqsQueueURL),
	})
	if err != nil {
		return fmt.Errorf("failed to send message to SQS: %v", err)
	}

	return nil
}

func main() {
	lambda.Start(processPaymentHandler)
}
