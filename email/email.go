package email

import "context"

type EmailService interface {
	SendTicketVerifyLink(ctx context.Context, email string, link string) error
}

type EmailMock struct {

}

// mocking object for testing always work 
func NewEmailMock() *EmailMock {
	return &EmailMock{}
}

func (m *EmailMock) SendTicketVerifyLink(ctx context.Context, email string, link string) error {
	return nil
}