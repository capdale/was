package localstorage

import (
	"context"
	"io"
	"path"

	"github.com/capdale/was/types/binaryuuid"
)

const collectionDirPath = "/collectionimg"

func (ls *LocalStorage) GetCollectionJPG(ctx context.Context, uuid binaryuuid.UUID) (*[]byte, error) {
	return nil, nil
}

func (ls *LocalStorage) UploadCollectionJPG(ctx context.Context, uuid binaryuuid.UUID, reader io.Reader) error {
	filepath := path.Join(collectionDirPath, uuid.String()+".jpg")
	if err := ls.upload(ctx, filepath, reader); err != nil {
		return err
	}
	return nil
}

func (ls *LocalStorage) DeleteCollectionJPG(ctx context.Context, uuid binaryuuid.UUID) error {
	filepath := path.Join(collectionDirPath, uuid.String()+".jpg")
	return ls.delete(ctx, filepath)
}
