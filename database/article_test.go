package database

import (
	"github.com/capdale/was/types/binaryuuid"
	"github.com/capdale/was/types/claimer"
	"github.com/stretchr/testify/assert"
)

func (s *DatabaseSuite) TestAnonymousCreateArticle() {
	var anonymousClaimer *claimer.Claimer = nil
	randomUUID, _ := binaryuuid.NewRandom()
	collections := &[]binaryuuid.UUID{randomUUID}
	order := &[]uint8{1}
	err := s.d.CreateNewArticle(anonymousClaimer, "title", "content", collections, collections, order)
	if !assert.NotNil(s.T(), err) {
		assert.ErrorIs(s.T(), err, ErrInvalidInput)
		return
	}
}
func (s *DatabaseSuite) TestCreateArticle() {
	s.createTestAccount("test@test.com", "test")
	claimer, _, err := s.loginAccount("test", "test")
	assert.Nil(s.T(), err)

	randomUUID, _ := binaryuuid.NewRandom()
	collections := &[]binaryuuid.UUID{randomUUID}
	order := &[]uint8{1}
	err = s.d.CreateNewArticle(claimer, "title", "content", collections, collections, order)
	assert.Nil(s.T(), err)
}
