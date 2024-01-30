package auth

import (
	"crypto/rand"
	"math/big"
)

const validLetters = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz-"

var tokenLength = big.NewInt(int64(len(validLetters)))

func GenerateRandomString(n int) (string, error) {
	ret := make([]byte, n)
	for i := 0; i < n; i++ {
		num, err := rand.Int(rand.Reader, tokenLength)
		if err != nil {
			return "", err
		}
		ret[i] = validLetters[num.Int64()]
	}

	return string(ret), nil
}

func RandToken(n int) (*[]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	return &b, err
}
