package utils

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
)

func AwsSession() (*session.Session, error) {
	// Create a new AWS session
	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String("us-west-1"),
		Credentials: credentials.NewStaticCredentials("AKIAR4GEI3NVFWDTGA5N", "QsnFuEPe+7NyuHyLvSR3kqqOZWlzSuk4k1jKhaOx", ""),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create new session: %v", err)
	}

	return sess, nil
}
