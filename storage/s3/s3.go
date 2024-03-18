package s3

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/credentials/ec2rolecreds"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/capdale/was/config"
	"github.com/capdale/was/storage"
)

var ErrInvalidCredForm = errors.New("invalid cred form")

type S3Bucket struct {
	client     *s3.Client
	bucketName *string
}

// must implement storage interface
var _ storage.Storage = (*S3Bucket)(nil)

func New(config *config.S3) (*S3Bucket, error) {
	var credProvider aws.CredentialsProvider
	if config.Id == nil && config.Key == nil {
		credProvider = ec2rolecreds.New()
	} else {
		if config.Id == nil || config.Key == nil {
			return nil, ErrInvalidCredForm
		}
		credProvider = credentials.NewStaticCredentialsProvider(*config.Id, *config.Key, "")
	}

	cfg, err := awsConfig.LoadDefaultConfig(context.TODO(), awsConfig.WithRegion(config.Region), awsConfig.WithCredentialsProvider(credProvider))
	if err != nil {
		return nil, err
	}
	s3Client := s3.NewFromConfig(cfg)
	return &S3Bucket{client: s3Client, bucketName: aws.String(config.Name)}, nil
}
