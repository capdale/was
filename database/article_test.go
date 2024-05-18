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
	account := s.MustCreateAccount()
	claimer := account.Claim

	randomUUID, _ := binaryuuid.NewRandom()
	collections := &[]binaryuuid.UUID{randomUUID}
	order := &[]uint8{1}
	err := s.d.CreateNewArticle(claimer, "title", "content", collections, collections, order)
	assert.Nil(s.T(), err)
}

func (s *DatabaseSuite) TestCommentArticle() {
	account := s.MustCreateAccount()
	claimer := account.Claim
	username := account.Username

	randomUUID, _ := binaryuuid.NewRandom()
	collections := &[]binaryuuid.UUID{randomUUID}
	order := &[]uint8{1}
	s.d.CreateNewArticle(claimer, "title", "content", collections, collections, order)
	linkIds, _ := s.d.GetArticleLinkIdsByUsername(claimer, &username, 0, 1)
	linkUUID := (*linkIds)[0]

	comment := "test comment"
	err := s.d.Comment(claimer, linkUUID, &comment)
	assert.Nil(s.T(), err)

	comments, err := s.d.GetComments(claimer, linkUUID, 0, 1)
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), "test comment", (*comments)[0].Comment)
}

func (s *DatabaseSuite) TestHeartArticle() {
	account := s.MustCreateAccount()
	claimer := account.Claim
	username := account.Username

	randomUUID, _ := binaryuuid.NewRandom()
	collections := &[]binaryuuid.UUID{randomUUID}
	order := &[]uint8{1}
	s.d.CreateNewArticle(claimer, "title", "content", collections, collections, order)
	linkIds, _ := s.d.GetArticleLinkIdsByUsername(claimer, &username, 0, 1)
	linkUUID := (*linkIds)[0]

	err := s.d.DoHeart(claimer, linkUUID, 1)
	assert.Nil(s.T(), err)

	article, _ := s.d.GetArticle(claimer, *linkUUID)
	assert.Equal(s.T(), uint64(1), article.Meta.HeartCount)

	err = s.d.DoHeart(claimer, linkUUID, 0)
	assert.Nil(s.T(), err)

	article, _ = s.d.GetArticle(claimer, *linkUUID)
	assert.Equal(s.T(), uint64(0), article.Meta.HeartCount)
}
