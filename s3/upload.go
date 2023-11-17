package s3

import (
	"bytes"
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func (s *BasicBucket) Upload(filename string, b *[]byte) error {
	reader := bytes.NewReader(*b)
	_, err := s.client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket: s.bucketName,
		Key:    aws.String(fmt.Sprintf("%s.jpg", filename)),
		Body:   reader,
	})
	return err
}
