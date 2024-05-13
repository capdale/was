package database

import (
	"github.com/capdale/was/types/binaryuuid"
	"github.com/capdale/was/types/claimer"
)

func (s *DatabaseSuite) createTestAccount(email string, username string) {
	ticket, _ := s.d.CreateTicketByEmail(email)
	s.d.CreateOriginViaTicket(ticket, username, "test")
}

func (s *DatabaseSuite) loginAccount(username string, password string) (*claimer.Claimer, *binaryuuid.UUID, error) {
	return s.d.GetOriginUserClaimerAndUUID(username, password)
}
