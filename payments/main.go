package main

import (
	"github.com/aws/aws-lambda-go/lambda"
)

func main() {
	handler := []string{"payments", "CreatePaymentHandler"}
	lambda.Start(handler)
}
