package email

import "context"

type EmailService interface {
	SendTicketVerifyLink(ctx context.Context, email string, link string) error
}
