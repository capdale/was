package s3

import (
	"context"
	"errors"
	"io"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"golang.org/x/sync/errgroup"
)

var ErrInvalidInput = errors.New("invalid input")

func (s *S3Bucket) get(ctx context.Context, filename string) (*s3.GetObjectOutput, error) {
	return s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: s.bucketName,
		Key:    aws.String(filename),
	})
}

func (s *S3Bucket) upload(ctx context.Context, filename string, reader io.Reader) (*s3.PutObjectOutput, error) {
	return s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: s.bucketName,
		Key:    aws.String(filename),
		Body:   reader,
	})
}

func (s *S3Bucket) delete(ctx context.Context, filename string) (*s3.DeleteObjectOutput, error) {
	return s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: s.bucketName,
		Key:    aws.String(filename),
	})
}

func (s *S3Bucket) uploadMultiple(ctx context.Context, filenames *[]string, readers *[]io.Reader) error {
	errGrp, errCtx := errgroup.WithContext(ctx)
	for i := 0; i < len(*filenames); i++ {
		i := i
		errGrp.Go(func() error { _, err := s.upload(errCtx, (*filenames)[i], (*readers)[i]); return err })
	}
	return errGrp.Wait()
}

func (s *S3Bucket) deleteMultiple(ctx context.Context, filenames *[]string) error {
	errGrp, errCtx := errgroup.WithContext(ctx)
	for i := 0; i < len(*filenames); i++ {
		i := i
		errGrp.Go(func() error { _, err := s.delete(errCtx, (*filenames)[i]); return err })
	}
	return errGrp.Wait()
}
