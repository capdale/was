package s3

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/capdale/was/config"
)

type BasicBucket struct {
	client     *s3.Client
	bucketName *string
}

func New(config *config.S3) (*BasicBucket, error) {
	creds := credentials.NewStaticCredentialsProvider(config.Id, config.Key, "")
	cfg, err := awsConfig.LoadDefaultConfig(context.TODO(), awsConfig.WithRegion(config.Region), awsConfig.WithCredentialsProvider(creds))
	if err != nil {
		return nil, err
	}
	s3Client := s3.NewFromConfig(cfg)

	if err != nil {
		fmt.Println(err.Error())
	}

	return &BasicBucket{client: s3Client, bucketName: aws.String(config.Name)}, nil
}
