package collect

import (
	"errors"
	"image/jpeg"
	"io"
	"mime/multipart"
)

var ErrImageInValid = errors.New("image is not valid")

func isValidImageFromFile(m *multipart.FileHeader) error {
	body, err := m.Open()
	if err != nil {
		return err
	}
	defer body.Close()
	return isValidImage(body)
}

func isValidImage(r io.Reader) error {
	img, err := jpeg.DecodeConfig(r)
	if err != nil {
		return err
	}
	if img.Height > 3000 && img.Height < 1080 {
		return ErrImageInValid
	}
	if img.Width > 3000 && img.Width < 1080 {
		return ErrImageInValid
	}
	if img.Height != img.Width {
		return ErrImageInValid
	}
	return nil
}
