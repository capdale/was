package originAPI

import (
	"errors"
	"unicode"
)

var ErrInvalidPasswordForm = errors.New("invalid password form")

func validatePassword(password *string) bool {
	var (
		hasUpper   = false
		hasLower   = false
		hasNumber  = false
		hasSpecial = false
	)
	passwordLen := len(*password)
	if passwordLen > 32 || passwordLen < 8 {
		return false
	}
	for _, char := range *password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsNumber(char):
			hasNumber = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}
	return hasUpper && hasLower && hasNumber && hasSpecial
}
