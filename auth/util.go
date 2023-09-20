package auth

import (
	"crypto/rand"
	"encoding/base64"
	"math/big"
)

const validLetters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ01234567890"

var tokenLength = big.NewInt(int64(len(validLetters)))

func createRandomToken() (string, error) {
	bytes := make([]byte, 20)
	for i := range bytes {
		num, err := rand.Int(rand.Reader, tokenLength)
		if err != nil {
			return "", err
		}
		bytes[i] = validLetters[num.Int64()]
	}
	return string(bytes), nil
}

func RandToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.StdEncoding.EncodeToString(b)
}
