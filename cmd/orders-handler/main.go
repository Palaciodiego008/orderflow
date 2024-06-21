package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/sqs"
)

type CreateOrderRequest struct {
	UserID     string `json:"user_id"`
	Item       string `json:"item"`
	Quantity   int    `json:"quantity"`
	TotalPrice int64  `json:"total_price"`
}

type CreateOrderEvent struct {
	OrderID    string `json:"order_id"`
	TotalPrice int64  `json:"total_price"`
}

var (
	dynamoDBTableName string
	sqsQueueURL       string
)

func init() {
	dynamoDBTableName = os.Getenv("DYNAMODB_TABLE_NAME")
	sqsQueueURL = os.Getenv("SQS_QUEUE_URL")

	if dynamoDBTableName == "" {
		dynamoDBTableName = "orders_table"
		fmt.Println("DYNAMODB_TABLE_NAME environment variable not set")
	}

	if sqsQueueURL == "" {
		sqsQueueURL = "https://sqs.us-east-1.amazonaws.com/129260641130/orders_queue"
		fmt.Println("SQS_QUEUE_URL environment variable not set")
	}
}

func createOrderHandler(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	var req CreateOrderRequest
	err := json.Unmarshal([]byte(request.Body), &req)
	if err != nil {
		return events.APIGatewayProxyResponse{StatusCode: 400, Body: "Invalid request body"}, nil
	}

	orderID := fmt.Sprintf("ORDER-%d", time.Now().UnixNano())

	err = storeOrderDetails(orderID, req)
	if err != nil {
		return events.APIGatewayProxyResponse{StatusCode: 500, Body: "Failed to store order details"}, nil
	}

	err = sendCreateOrderEventSQS(orderID, req.TotalPrice)
	if err != nil {
		return events.APIGatewayProxyResponse{StatusCode: 500, Body: "Failed to send create order event"}, nil
	}

	responseBody := fmt.Sprintf("Order created successfully. Order ID: %s", orderID)
	return events.APIGatewayProxyResponse{StatusCode: 200, Body: responseBody}, nil
}

func storeOrderDetails(orderID string, req CreateOrderRequest) error {
	sess := session.Must(session.NewSession())
	svc := dynamodb.New(sess)

	input := &dynamodb.PutItemInput{
		TableName: aws.String(dynamoDBTableName),
		Item: map[string]*dynamodb.AttributeValue{
			"PK":          {S: aws.String(fmt.Sprintf("ORDER#%s", orderID))},
			"SK":          {S: aws.String("DETAILS")},
			"order_id":    {S: aws.String(orderID)},
			"user_id":     {S: aws.String(req.UserID)},
			"item":        {S: aws.String(req.Item)},
			"quantity":    {N: aws.String(fmt.Sprintf("%d", req.Quantity))},
			"total_price": {N: aws.String(fmt.Sprintf("%d", req.TotalPrice))},
		},
	}

	_, err := svc.PutItem(input)
	return err
}

func sendCreateOrderEventSQS(orderID string, totalPrice int64) error {
	sess := session.Must(session.NewSession())
	sqsSvc := sqs.New(sess)

	event := map[string]interface{}{
		"order_id":    orderID,
		"total_price": totalPrice,
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
	lambda.Start(createOrderHandler)
}
