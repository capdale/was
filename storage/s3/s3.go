package s3

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/capdale/was/config"
	"github.com/capdale/was/storage"
)

type S3Bucket struct {
	client     *s3.Client
	bucketName *string
}

// must implement storage interface
var _ storage.Storage = (*S3Bucket)(nil)

func New(config *config.S3) (*S3Bucket, error) {
	creds := credentials.NewStaticCredentialsProvider(config.Id, config.Key, "")
	cfg, err := awsConfig.LoadDefaultConfig(context.TODO(), awsConfig.WithRegion(config.Region), awsConfig.WithCredentialsProvider(creds))
	if err != nil {
		return nil, err
	}
	s3Client := s3.NewFromConfig(cfg)

	if err != nil {
		fmt.Println(err.Error())
	}

	return &S3Bucket{client: s3Client, bucketName: aws.String(config.Name)}, nil
}
