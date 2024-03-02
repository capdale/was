package localstorage

import (
	"context"
	"io"
	"path"

	"github.com/capdale/was/storage"
	"github.com/capdale/was/types/binaryuuid"
)

const articleDirPath = "/articleimg"

func (ls *LocalStorage) UploadArticleJPGs(ctx context.Context, uuids *[]binaryuuid.UUID, readers *[]io.Reader) error {
	if len(*uuids) != len(*readers) {
		return storage.ErrInvalidInput
	}
	filepaths := make([]string, len(*uuids))

	for i, uuid := range *uuids {
		filepaths[i] = path.Join(articleDirPath, uuid.String()+".jpg")
	}

	if err := ls.uploadMultiple(ctx, &filepaths, readers); err != nil {
		return err
	}
	return nil
}

func (ls *LocalStorage) DeleteArticleJPGs(ctx context.Context, uuids *[]binaryuuid.UUID) error {
	filepaths := make([]string, len(*uuids))
	for i, uuid := range *uuids {
		filepaths[i] = path.Join(articleDirPath, uuid.String()+".jpg")
	}

	if err := ls.deleteMultiple(ctx, &filepaths); err != nil {
		return err
	}
	return nil
}
