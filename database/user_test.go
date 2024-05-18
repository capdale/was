package database

import (
	"github.com/capdale/was/types/claimer"
)

func (s *DatabaseSuite) createTestAccount(email string, username string, password string) {
	ticket, _ := s.d.CreateTicketByEmail(email)
	s.d.CreateOriginViaTicket(ticket, username, password)
}

func (s *DatabaseSuite) loginAccount(username string, password string) (*claimer.Claimer, error) {
	return s.d.GetOriginUserClaim(username, password)
}
