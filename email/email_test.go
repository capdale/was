package email_test

import (
	"testing"

	"github.com/capdale/was/email"
)

func TestCensorEmail(t *testing.T) {
	email.CensorEmail("")
}
