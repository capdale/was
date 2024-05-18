package email

import (
	"bytes"
	"context"
	"fmt"
	"regexp"

	"github.com/capdale/was/config"
)

type EmailService interface {
	SendTicketVerifyLink(ctx context.Context, email string, link string) error
}

type EmailMock struct {
	logtype string
}

// mocking object for testing always work
func NewEmailMock(c *config.Mock) *EmailMock {
	return &EmailMock{
		logtype: c.Type,
	}
}

func (m *EmailMock) SendTicketVerifyLink(ctx context.Context, email string, link string) error {
	if m.logtype == "cli" {
		fmt.Println(link)
	}
	return nil
}

var emailCensorExpr = regexp.MustCompile(`^[\w-\.]([\w-\.]*)@([\w-])([\w-]*)\.([\w-]+\.)*([\w-])([\w-]{1,3})$`)

func CensorEmail(email string) string {
	subMatches := emailCensorExpr.FindSubmatchIndex([]byte(email))
	emailLength := len(email)
	censoredEmail := bytes.Repeat([]byte("*"), emailLength)
	for i := 2; i < len(subMatches); i += 2 {
		front := subMatches[i] - 1
		if front < 0 {
			continue
		}
		censoredEmail[front] = byte(email[front])
	}
	return string(censoredEmail)
}
