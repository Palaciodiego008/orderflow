# Order-Payments CDK Go Project

This project is a serverless application built with AWS CDK (Cloud Development Kit) using Go. It includes two microservices: one for processing orders and another for processing payments. These microservices communicate via SQS and use DynamoDB for data storage. The services are deployed on AWS Lambda and are triggered via API Gateway.

## Project Structure

The project contains the following components:

- **Order Service**: Handles creating orders and sending payment events.
- **Payment Service**: Handles receiving payments and updating order statuses.

## Endpoints

### Order Service

- **POST /orders**

  **Request Body**:
  ```json
  {
    "user_id": "string",
    "item": "string",
    "quantity": int,
    "total_price": int64
  }
  ```

  **Response**:
  - 200 OK if the order is created successfully.
  - 400 Bad Request if the request body is invalid.
  - 500 Internal Server Error if there is an issue creating the order.

### Payment Service

- **POST /payments**

  **Request Body**:
  ```json
  {
    "order_id": "string",
    "status": "string"
  }
  ```

  **Response**:
  - 200 OK if the payment is processed successfully.
  - 400 Bad Request if the request body is invalid.
  - 500 Internal Server Error if there is an issue processing the payment.

## Deployment

### Prerequisites

Ensure you have the following installed:

- AWS CLI
- AWS CDK Toolkit
- Go

### Steps to Deploy

1. **Configure AWS CLI**: Make sure your AWS CLI is configured with the appropriate credentials.
   ```bash
   aws configure
   ```

2. **Initialize CDK Project**: If you haven't already, initialize a new CDK project.
   ```bash
   cdk init app --language go
   ```

3. **Set Up Project Structure**: Ensure your project structure looks like this:
   ```
   order-payments/
   ├── cmd/
   │   ├── orders-handler/
   │   │   └── main.go
   │   ├── payments-handler/
   │   │   └── main.go
   ├── lib/
   │   └── order_payments_stack.go
   ├── go.mod
   ├── go.sum
   ├── cdk.json
   └── main.go
   ```

4. **Install Dependencies**: Install the necessary dependencies.
   ```bash
   go mod tidy
   ```

5. **Build Project**: Compile the Go project to ensure there are no errors.
   ```bash
   go build ./...
   ```

6. **Deploy with CDK**: Use CDK to deploy the stack.
   ```bash
   cdk deploy
   ```

## Useful Commands

- `cdk deploy` - Deploy this stack to your default AWS account/region.
- `cdk diff` - Compare deployed stack with current state.
- `cdk synth` - Emits the synthesized CloudFormation template.
- `go test` - Run unit tests.

## Summary

This project demonstrates how to build and deploy a serverless application with AWS CDK using Go. It includes microservices for order processing and payment processing, using API Gateway, Lambda, SQS, and DynamoDB. Follow the steps above to deploy and manage your stack in AWS.
