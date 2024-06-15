package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/google/uuid"
)

var (
	dynamoDbClient *dynamodb.DynamoDB
	sqsClient      *sqs.SQS
	tableName      = os.Getenv("DYNAMO_DB_TABLE")
	queueUrl       = os.Getenv("SQS_QUEUE_URL")
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

func init() {
	sess := session.Must(session.NewSession())
	dynamoDbClient = dynamodb.New(sess)
	sqsClient = sqs.New(sess)
}

func CreateOrder(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	var orderRequest CreateOrderRequest
	err := json.Unmarshal([]byte(req.Body), &orderRequest)
	if err != nil {
		return events.APIGatewayProxyResponse{StatusCode: http.StatusBadRequest}, err
	}

	orderID := uuid.New().String()
	orderItem := map[string]*dynamodb.AttributeValue{
		"OrderID": {
			S: aws.String(orderID),
		},
		"UserID": {
			S: aws.String(orderRequest.UserID),
		},
		"Item": {
			S: aws.String(orderRequest.Item),
		},
		"Quantity": {
			N: aws.String(fmt.Sprintf("%d", orderRequest.Quantity)),
		},
		"TotalPrice": {
			N: aws.String(fmt.Sprintf("%d", orderRequest.TotalPrice)),
		},
		"OrderStatus": {
			S: aws.String("INCOMPLETE"),
		},
		"CreatedAt": {
			S: aws.String(time.Now().Format(time.RFC3339)),
		},
	}

	_, err = dynamoDbClient.PutItem(&dynamodb.PutItemInput{
		TableName: aws.String(tableName),
		Item:      orderItem,
	})
	if err != nil {
		return events.APIGatewayProxyResponse{StatusCode: http.StatusInternalServerError}, err
	}

	orderEvent := CreateOrderEvent{
		OrderID:    orderID,
		TotalPrice: orderRequest.TotalPrice,
	}

	eventBody, err := json.Marshal(orderEvent)
	if err != nil {
		return events.APIGatewayProxyResponse{StatusCode: http.StatusInternalServerError}, err
	}

	_, err = sqsClient.SendMessage(&sqs.SendMessageInput{
		QueueUrl:    aws.String(queueUrl),
		MessageBody: aws.String(string(eventBody)),
	})
	if err != nil {
		return events.APIGatewayProxyResponse{StatusCode: http.StatusInternalServerError}, err
	}

	return events.APIGatewayProxyResponse{StatusCode: http.StatusCreated}, nil
}
