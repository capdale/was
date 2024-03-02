package s3

import (
	"context"
	"fmt"
	"io"

	"github.com/capdale/was/types/binaryuuid"
)

const articleImageFmt = `article-img-%s.jpg`

func (s *S3Bucket) UploadArticleJPGs(ctx context.Context, uuids *[]binaryuuid.UUID, readers *[]io.Reader) error {
	count := len(*uuids)
	if count != len(*readers) {
		return ErrInvalidInput
	}

	filenames := make([]string, count)
	for i, uuid := range *uuids {
		filenames[i] = fmt.Sprintf(articleImageFmt, uuid.String())
	}

	return s.uploadMultiple(ctx, &filenames, readers)
}

func (s *S3Bucket) DeleteArticleJPGs(ctx context.Context, uuids *[]binaryuuid.UUID) error {
	filenames := make([]string, len(*uuids))
	for i, uuid := range *uuids {
		filenames[i] = fmt.Sprintf(articleImageFmt, uuid.String())
	}

	return s.deleteMultiple(ctx, &filenames)
}
