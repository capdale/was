package s3

import (
	"context"
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func (s *BasicBucket) UploadJPG(ctx context.Context, filename string, reader io.Reader) (*s3.PutObjectOutput, error) {
	return s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: s.bucketName,
		Key:    aws.String(fmt.Sprintf("%s.jpg", filename)),
		Body:   reader,
	})
}

func (s *BasicBucket) DeleteJPG(filename string) (*s3.DeleteObjectOutput, error) {
	return s.client.DeleteObject(context.TODO(), &s3.DeleteObjectInput{
		Bucket: s.bucketName,
		Key:    aws.String(fmt.Sprintf("%s.jpg", filename)),
	})
}
