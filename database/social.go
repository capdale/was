package database

import (
	"errors"

	"github.com/capdale/was/model"
	"github.com/capdale/was/types/binaryuuid"
	"gorm.io/gorm"
)

func (d *DB) HasQueryPermission(claimerId int64, targetId int64) (bool, error) {
	if claimerId == targetId {
		return true, nil
	}

	userDisplayType := &model.UserDisplayType{}
	if err := d.DB.
		Select("is_private").
		Where("user_id = ?", targetId).
		First(userDisplayType).Error; err != nil {
		return false, err
	}
	if !userDisplayType.IsPrivate {
		return true, nil
	}

	if claimerId == -1 {
		return false, nil
	}

	var exist bool

	err := d.DB.
		Model(&model.UserFollow{}).
		Select("count(*) > 0").
		Where("user_id = ? AND target_id = ?", claimerId, targetId).
		Find(&exist).Error

	if err != nil {
		return false, err
	}
	return exist, nil
}

func (d *DB) RequestFollow(claimer binaryuuid.UUID, target string) error {

	claimerId, err := d.GetUserIdByUUID(claimer)
	if err != nil {
		return err
	}

	targetId, err := d.GetUserIdByName(target)
	if err != nil {
		return err
	}

	if claimerId == targetId {
		return ErrInvalidInput
	}

	isPublic, err := d.IsUserPublic(targetId)
	if err != nil {
		return err
	}

	// if target user is public account, then follow is done immediately
	if isPublic {
		return d.followUser(claimerId, targetId)
	}

	// if target user is private account, then request follow
	return d.DB.Create(&model.UserFollowRequest{
		UserId:   claimerId,
		TargetId: targetId,
	}).Error
}

func (d *DB) followUser(followerId int64, followingId int64) error {
	return d.DB.Create(&model.UserFollow{
		UserId:   followerId,
		TargetId: followingId,
	}).Error
}

func (d *DB) IsUserPublic(userId int64) (bool, error) {
	display := &model.UserDisplayType{}
	if err := d.DB.
		Select("is_private").
		Where("user_id = ?", userId).
		First(&display).Error; err != nil {
		return false, nil
	}
	return !display.IsPrivate, nil
}

func (d *DB) IsFollower(claimerUUID binaryuuid.UUID, targetname string) (isFollower bool, err error) {
	claimerId, err := d.GetUserIdByUUID(claimerUUID)
	if err != nil {
		return
	}
	targetId, err := d.GetUserIdByName(targetname)
	if err != nil {
		return
	}
	r := &model.UserFollow{}
	if err = d.DB.
		Select("").
		Where("user_id = ? AND target_id = ?", claimerId, targetId).
		First(r).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return
	}
	return true, nil
}

func (d *DB) IsFollowing(claimerUUID binaryuuid.UUID, targetname string) (bool, error) {
	claimerId, err := d.GetUserIdByUUID(claimerUUID)
	if err != nil {
		return false, err
	}
	targetId, err := d.GetUserIdByName(targetname)
	if err != nil {
		return false, err
	}
	r := &model.UserFollow{}
	if err = d.DB.
		Select("").
		Where("user_id = ? AND target_id = ?", targetId, claimerId).
		First(r).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (d *DB) GetFollowers(username string, offset int, limit int) (*[]string, error) {
	userId, err := d.GetUserIdByName(username)
	if err != nil {
		return nil, err
	}

	followerNames := []string{}
	if err := d.DB.
		Model(&model.User{}).
		Select("users.username").
		Joins("JOIN user_follows ON user_follows.target_id = ? AND user_follows.user_id = users.id", userId).
		Offset(offset).
		Limit(limit).
		Find(&followerNames).Error; err != nil {
		return nil, err
	}
	return &followerNames, nil
}

func (d *DB) GetFollowings(username string, offset int, limit int) (*[]string, error) {
	userId, err := d.GetUserIdByName(username)
	if err != nil {
		return nil, err
	}

	followingNames := []string{}
	if err := d.DB.
		Model(&model.User{}).
		Select("users.username").
		Joins("JOIN user_follows ON user_follows.user_id = ? AND user_follows.target_id = users.id", userId).
		Offset(offset).
		Limit(limit).
		Find(&followingNames).Error; err != nil {
		return nil, err
	}
	return &followingNames, nil
}

func (d *DB) AcceptRequestFollow(claimerUUID *binaryuuid.UUID, requestUUID *binaryuuid.UUID) error {
	claimerId, err := d.GetUserIdByUUID(*claimerUUID)
	if err != nil {
		return err
	}
	requestId, err := d.GetUserIdByUUID(*requestUUID)
	if err != nil {
		return err
	}

	request := &model.UserFollowRequest{}
	if err := d.DB.
		Select("id").
		Where("user_id = ? and target_id = ?", claimerId, requestId).
		First(request).Error; err != nil {
		return err
	}

	if err := d.DB.Delete(request).Error; err != nil {
		return err
	}

	return d.followUser(claimerId, requestId)
}

func (d *DB) RemoveFollower(claimerUUID *binaryuuid.UUID, targetname string) error {
	claimerId, err := d.GetUserIdByUUID(*claimerUUID)
	if err != nil {
		return err
	}

	targetId, err := d.GetUserIdByName(targetname)
	if err != nil {
		return err
	}

	result := d.DB.
		Where("user_id = ? AND target_id = ?", targetId, claimerId).
		Delete(&model.UserFollow{})
	if result.RowsAffected < 1 {
		return ErrNoAffectedRow
	}
	return result.Error
}

func (d *DB) RemoveFollowing(claimerUUID *binaryuuid.UUID, targetname string) error {
	claimerId, err := d.GetUserIdByUUID(*claimerUUID)
	if err != nil {
		return err
	}

	targetId, err := d.GetUserIdByName(targetname)
	if err != nil {
		return err
	}

	result := d.DB.
		Where("user_id = ? AND target_id = ?", claimerId, targetId).
		Delete(&model.UserFollow{})
	if result.RowsAffected < 1 {
		return ErrNoAffectedRow
	}
	return result.Error
}
