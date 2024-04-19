package articleAPI

import (
	"encoding/base64"
	"errors"

	"github.com/capdale/was/types/binaryuuid"
)

var ErrInvalidLink = errors.New("invalid link form")

func decodeLink(link string) (*binaryuuid.UUID, error) {
	linkBytes, err := base64.URLEncoding.DecodeString(link)
	if err != nil {
		return nil, err
	}

	if len(linkBytes) != 16 {
		return nil, ErrInvalidLink
	}

	linkId, err := binaryuuid.FromBytes(linkBytes)
	return &linkId, err
}
