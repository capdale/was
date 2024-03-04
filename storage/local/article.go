package localstorage

import (
	"context"
	"io"
	"os"
	"path"

	"github.com/capdale/was/storage"
	"github.com/capdale/was/types/binaryuuid"
)

const articleDirPath = "/articleimg"

func (ls *LocalStorage) GetArticleJPG(ctx context.Context, uuid binaryuuid.UUID) (*[]byte, error) {
	filepath := path.Join(ls.baseDir, articleDirPath, uuid.String()+".jpg")
	if _, err := os.Stat(filepath); err != nil {
		return nil, err
	}
	file, err := os.ReadFile(filepath)
	if err != nil {
		return nil, err
	}
	return &file, nil
}

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
