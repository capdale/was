package database

import (
	"github.com/stretchr/testify/assert"
)

func (s *DatabaseSuite) TestAcceptanceSocialFollow() {
	user1 := s.MustCreateAccount()
	user2 := s.MustCreateAccount()
	user3 := s.MustCreateAccount()

	// default user visibility type is public
	// so, request follow immediately accepted
	// user1 follow user2
	err := s.d.RequestFollow(user1.Claim, &user2.Username)
	assert.Nil(s.T(), err)

	// check user2 followers [user1]
	followers, _ := s.d.GetFollowers(user1.Claim, &user2.Username, 0, 1)
	assert.Equal(s.T(), user1.Username, (*followers)[0])

	// check user2 followings []
	followings, _ := s.d.GetFollowings(user1.Claim, &user2.Username, 0, 2)
	assert.Len(s.T(), (*followings), 0)

	// check user1 followings [user2]
	followings, _ = s.d.GetFollowings(user1.Claim, &user1.Username, 0, 1)
	assert.Equal(s.T(), user2.Username, (*followings)[0])

	// user2 follow user1
	err = s.d.RequestFollow(user2.Claim, &user1.Username)
	assert.Nil(s.T(), err)

	// check user2 followers [user1]
	followers, _ = s.d.GetFollowers(user1.Claim, &user2.Username, 0, 1)
	assert.Equal(s.T(), user1.Username, (*followers)[0])

	// check user2 followings [user1]
	followings, _ = s.d.GetFollowings(user1.Claim, &user2.Username, 0, 1)
	assert.Equal(s.T(), user1.Username, (*followings)[0])

	// change user3 visibility to private
	s.d.ChangeVisibility(user3.Claim, userVisibilityPrivate)

	// user1 follow user3
	s.d.RequestFollow(user1.Claim, &user3.Username)

	// check user3 followers []
	followers, _ = s.d.GetFollowers(user3.Claim, &user3.Username, 0, 1)
	assert.Len(s.T(), *followers, 0)

	// check user3 follow requests [user1]
	requests, _ := s.d.GetFollowRequests(user3.Claim, 0, 4)

	// accept follow request
	s.d.AcceptRequestFollow(user3.Claim, &((*requests)[0].Code))

	// check user1 follow user3
	followings, _ = s.d.GetFollowings(user1.Claim, &user1.Username, 0, 2)
	assert.Contains(s.T(), *followings, user3.Username)

	// check user3 follower [user1]
	followers, _ = s.d.GetFollowers(user3.Claim, &user3.Username, 0, 1)
	assert.Equal(s.T(), user1.Username, (*followers)[0])
}
