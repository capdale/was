package storage

import (
	"context"
	"errors"
	"io"

	"github.com/capdale/was/types/binaryuuid"
)

var ErrInvalidInput = errors.New("invalid input")

type Storage interface {
	GetCollectionJPG(ctx context.Context, uuid binaryuuid.UUID) (*[]byte, error)
	UploadCollectionJPG(ctx context.Context, uuid binaryuuid.UUID, reader io.Reader) error
	DeleteCollectionJPG(ctx context.Context, uuid binaryuuid.UUID) error
	GetArticleJPG(ctx context.Context, uuid binaryuuid.UUID) (*[]byte, error)
	UploadArticleJPGs(ctx context.Context, uuids *[]binaryuuid.UUID, readers *[]io.Reader) error
	DeleteArticleJPGs(ctx context.Context, uuids *[]binaryuuid.UUID) error
}
