package database

import (
	"github.com/stretchr/testify/assert"
)

func (s *DatabaseSuite) TestAcceptanceSocialFollow() {
	s.createTestAccount("test1@test.com", "test1", "test")
	s.createTestAccount("test2@test.com", "test2", "test")
	s.createTestAccount("test3@test.com", "test3", "test")

	claimer1, user1, _ := s.loginAccount("test1", "test")
	claimer2, user2, _ := s.loginAccount("test2", "test")
	claimer3, user3, _ := s.loginAccount("test3", "test")

	// default user visibility type is public
	// so, request follow immediately accepted
	// user1 follow user2
	err := s.d.RequestFollow(claimer1, user2)
	assert.Nil(s.T(), err)

	// check user2 followers [user1]
	followers, _ := s.d.GetFollowers(claimer1, user2, 0, 1)
	assert.Equal(s.T(), *user1, *((*followers)[0]))

	// check user2 followings []
	followings, _ := s.d.GetFollowings(claimer1, user2, 0, 2)
	assert.Len(s.T(), (*followings), 0)

	// check user1 followings [user2]
	followings, _ = s.d.GetFollowings(claimer1, user1, 0, 1)
	assert.Equal(s.T(), *user2, *((*followings)[0]))

	// user2 follow user1
	err = s.d.RequestFollow(claimer2, user1)
	assert.Nil(s.T(), err)

	// check user2 followers [user1]
	followers, _ = s.d.GetFollowers(claimer1, user2, 0, 1)
	assert.Equal(s.T(), *user1, *((*followers)[0]))

	// check user2 followings [user1]
	followings, _ = s.d.GetFollowings(claimer1, user2, 0, 1)
	assert.Equal(s.T(), *user1, *((*followings)[0]))

	// change user3 visibility to private
	s.d.ChangeVisibility(claimer3, userVisibilityPrivate)

	// user1 follow user3
	s.d.RequestFollow(claimer1, user3)

	// check user3 followers []
	followers, _ = s.d.GetFollowers(claimer3, user3, 0, 1)
	assert.Len(s.T(), *followers, 0)

	// check user3 follow requests [user1]
	requests, _ := s.d.GetFollowRequests(claimer3, 0, 4)
	assert.Equal(s.T(), user1.String(), (*requests)[0].String())

	// accept follow request
	s.d.AcceptRequestFollow(claimer3, user1)

	// check user1 follow user3
	followings, _ = s.d.GetFollowings(claimer1, user1, 0, 2)
	assert.Equal(s.T(), *user3, *((*followings)[1]))

	// check user3 follower [user1]
	requests, _ = s.d.GetFollowers(claimer3, user3, 0, 1)
	assert.Equal(s.T(), *user1, *((*requests)[0]))
}
