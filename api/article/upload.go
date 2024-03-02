package articleAPI

import (
	"context"
	"io"
	"mime/multipart"

	"github.com/capdale/was/types/binaryuuid"
)

func (a *ArticleAPI) uploadImagesWithUUID(ctx context.Context, uuids *[]binaryuuid.UUID, files *[]*multipart.FileHeader) error {
	readers := make([]io.Reader, len(*uuids))
	for i, fileHeader := range *files {
		r, err := fileHeader.Open()
		if err != nil {
			return err
		}
		defer r.Close()
		readers[i] = r
	}
	err := a.Storage.UploadArticleJPGs(ctx, uuids, &readers)
	return err
}
