package localstorage

import (
	"context"
	"io"
	"os"
	"path"

	"github.com/capdale/was/storage"
	"golang.org/x/sync/errgroup"
)

type LocalStorage struct {
	baseDir string
}

var _ storage.Storage = (*LocalStorage)(nil)

func New(baseDir string) (*LocalStorage, error) {
	if err := initializeLocalStorage(baseDir); err != nil {
		return nil, err
	}
	return &LocalStorage{
		baseDir: baseDir,
	}, nil
}

func initializeLocalStorage(baseDir string) (err error) {
	if err = os.Mkdir(path.Join(baseDir, collectionDirPath), os.ModePerm); err != nil {
		if !os.IsExist(err) {
			return
		}
	}
	if err = os.Mkdir(path.Join(baseDir, articleDirPath), os.ModePerm); err != nil {
		if !os.IsExist(err) {
			return
		}
	}
	return nil
}

func (ls *LocalStorage) upload(ctx context.Context, filepath string, reader io.Reader) error {
	filepath = path.Join(ls.baseDir, filepath)
	f, err := os.OpenFile(filepath, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.ReadFrom(reader)
	if err != nil {
		return err
	}
	return nil
}

func (ls *LocalStorage) uploadMultiple(ctx context.Context, filenames *[]string, readers *[]io.Reader) error {
	errGrp, errCtx := errgroup.WithContext(ctx)
	for i := 0; i < len(*filenames); i++ {
		i := i
		errGrp.Go(func() error { return ls.upload(errCtx, (*filenames)[i], (*readers)[i]) })
	}
	return errGrp.Wait()
}

func (ls *LocalStorage) delete(ctx context.Context, filepath string) error {
	filepath = path.Join(ls.baseDir, filepath)
	return os.Remove(filepath)
}

func (ls *LocalStorage) deleteMultiple(ctx context.Context, filepaths *[]string) error {
	errGrp, errCtx := errgroup.WithContext(ctx)
	for _, filepath := range *filepaths {
		errGrp.Go(func() error { return ls.delete(errCtx, filepath) })
	}
	return errGrp.Wait()
}
