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
	dynamoDBTableName = os.Getenv("DYNAMODB_TABLE_NAME")
	sqsQueueURL = os.Getenv("SQS_QUEUE_URL")

	if dynamoDBTableName == "" {
		fmt.Println("DYNAMODB_TABLE_NAME environment variable not set")
		dynamoDBTableName = "payments_table"
	}

	if sqsQueueURL == "" {
		fmt.Println("SQS_QUEUE_URL environment variable not set")
		sqsQueueURL = "https://sqs.us-east-1.amazonaws.com/129260641130/payments_queue"
	}
}

func processPaymentHandler(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	var req ProcessPaymentRequest
	err := json.Unmarshal([]byte(request.Body), &req)
	if err != nil {
		return events.APIGatewayProxyResponse{StatusCode: 400, Body: "Invalid request body"}, nil
	}

	log.Printf("Processing payment for order ID: %s, Status: %s", req.OrderID, req.Status)

	err = updatePaymentStatus(req.OrderID, req.Status)
	if err != nil {
		return events.APIGatewayProxyResponse{StatusCode: 500, Body: "Failed to update payment status"}, nil
	}

	err = sendOrderCompletedEvent(req.OrderID)
	if err != nil {
		return events.APIGatewayProxyResponse{StatusCode: 500, Body: "Failed to send order completed event"}, nil
	}

	responseBody := fmt.Sprintf("Payment processed successfully for order ID: %s", req.OrderID)
	return events.APIGatewayProxyResponse{StatusCode: 200, Body: responseBody}, nil
}

func updatePaymentStatus(orderID string, status string) error {
	sess := session.Must(session.NewSession())
	svc := dynamodb.New(sess)

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

	_, err := svc.UpdateItem(input)
	return err
}

func sendOrderCompletedEvent(orderID string) error {
	sess := session.Must(session.NewSession())
	sqsSvc := sqs.New(sess)

	event := OrderCompletedEvent{
		OrderID: orderID,
	}
	eventJSON, err := json.Marshal(event)
	if err != nil {
		return err
	}

	_, err = sqsSvc.SendMessage(&sqs.SendMessageInput{
		MessageBody: aws.String(string(eventJSON)),
		QueueUrl:    aws.String(sqsQueueURL),
	})
	return err
}

func main() {
	lambda.Start(processPaymentHandler)
}
