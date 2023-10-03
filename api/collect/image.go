package collect

import (
	"bytes"
	"image/jpeg"
	"io"

	"github.com/gin-gonic/gin"
)

func isValidImage(b *[]byte) bool {
	img, err := jpeg.DecodeConfig(bytes.NewReader(*b))
	if err != nil {
		return false
	}
	if img.Height > 3000 && img.Height < 1080 {
		return false
	}
	if img.Width > 3000 && img.Width < 1080 {
		return false
	}
	if img.Height != img.Width {
		return false
	}
	return true
}

func getByteFromCTX(ctx *gin.Context) (*[]byte, error) {
	body := ctx.Request.Body
	defer body.Close()

	b, err := io.ReadAll(body)
	if err != nil {
		return nil, err
	}
	return &b, nil
}
