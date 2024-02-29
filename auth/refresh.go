package auth

import (
	"github.com/capdale/was/types/binaryuuid"
)

// generate token uid and refresh random token
func (a *Auth) generateRefreshToken() (*binaryuuid.UUID, *[]byte, error) {
	randomUUID, err := binaryuuid.NewRandom()
	if err != nil {
		return nil, nil, err
	}
	randBytes, err := RandToken(48)
	if err != nil {
		return nil, nil, err
	}
	return &randomUUID, randBytes, nil
}
