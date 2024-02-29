package articleAPI

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"

	"github.com/capdale/was/types/binaryuuid"
)

const articleImageFmt = `article-img-%s`

func (a *ArticleAPI) uploadImagesWithUUID(ctx context.Context, uuids *[]binaryuuid.UUID, files *[]*multipart.FileHeader) error {
	count := len(*uuids)
	if count != len(*files) {
		return ErrInvalidForm
	}
	filenames := make([]string, count)
	readers := make([]io.Reader, count)
	for i := 0; i < count; i++ {
		filenames[i] = fmt.Sprintf(articleImageFmt, (*uuids)[i].String())
		r, err := (*files)[i].Open()
		if err != nil {
			return err
		}
		defer r.Close()
		readers = append(readers, r)
	}
	err := a.Storage.UploadJPGs(ctx, &filenames, &readers)
	return err
}
