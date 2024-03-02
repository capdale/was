package s3

import (
	"context"
	"fmt"
	"io"

	"github.com/capdale/was/types/binaryuuid"
)

const (
	collectionImageFmt = `collection-img-%s.jpg`
)

func (s *S3Bucket) GetCollectionJPG(ctx context.Context, uuid binaryuuid.UUID) (*[]byte, error) {
	filename := fmt.Sprintf(collectionImageFmt, uuid)
	output, err := s.get(ctx, filename)
	if err != nil {
		return nil, err
	}
	bytes, err := io.ReadAll(output.Body)
	if err != nil {
		return nil, err
	}
	defer output.Body.Close()

	return &bytes, nil
}

func (s *S3Bucket) UploadCollectionJPG(ctx context.Context, uuid binaryuuid.UUID, reader io.Reader) error {
	filename := fmt.Sprintf(collectionImageFmt, uuid)
	_, err := s.upload(ctx, filename, reader)
	return err
}

func (s *S3Bucket) DeleteCollectionJPG(ctx context.Context, uuid binaryuuid.UUID) error {
	filename := fmt.Sprintf(collectionImageFmt, uuid)
	_, err := s.delete(ctx, filename)
	return err
}
