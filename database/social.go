package database

import (
	"errors"

	"github.com/capdale/was/model"
	"github.com/capdale/was/types/binaryuuid"
	"github.com/capdale/was/types/claimer"
	"gorm.io/gorm"
)

func hasQueryPermission(tx *gorm.DB, claimerId uint64, targetId uint64) (bool, error) {
	if claimerId == targetId {
		return true, nil
	}
	var exist bool = false
	err := tx.Transaction(func(tx *gorm.DB) error {
		userDisplayType := &model.UserDisplayType{}
		if err := tx.
			Select("is_private").
			Where("user_id = ?", targetId).
			First(userDisplayType).Error; err != nil {
			return err
		}
		if !userDisplayType.IsPrivate {
			exist = true
			return nil
		}
		return tx.
			Model(&model.UserFollow{}).
			Select("count(*) > 0").
			Where("user_id = ? AND target_id = ?", claimerId, targetId).
			Find(&exist).Error
	})
	return exist, err
}

func (d *DB) RequestFollow(claimer *claimer.Claimer, targetname *string) error {
	return d.DB.Transaction(func(tx *gorm.DB) error {
		claimerId, err := getUserIdByClaimer(tx, claimer)
		if err != nil {
			return err
		}

		targetId, err := getUserIdByName(tx, targetname)
		if err != nil {
			return err
		}

		if claimerId == targetId {
			return ErrInvalidInput
		}
		isPublic, err := isUserPublic(tx, targetId)
		if err != nil {
			return err
		}

		// if target user is public account, then follow is done immediately
		if isPublic {
			return followUser(tx, claimerId, targetId)
		}

		// if target user is private account, then request follow
		return tx.Create(&model.UserFollowRequest{
			UserId:   claimerId,
			TargetId: targetId,
		}).Error
	})
}

func followUser(tx *gorm.DB, followerId uint64, followingId uint64) error {
	return tx.Create(&model.UserFollow{
		UserId:   followerId,
		TargetId: followingId,
	}).Error
}

func (d *DB) IsUserPublic(username *string) (bool, error) {
	var public = false
	err := d.DB.Transaction(func(tx *gorm.DB) error {
		var err error
		userId, err := getUserIdByName(tx, username)
		if err != nil {
			return err
		}
		public, err = isUserPublic(tx, userId)
		return err
	})
	return public, err
}

func isUserPublic(tx *gorm.DB, userId uint64) (bool, error) {
	display := &model.UserDisplayType{}
	if err := tx.
		Select("is_private").
		Where("user_id = ?", userId).
		First(&display).Error; err != nil {
		return false, nil
	}
	return !display.IsPrivate, nil
}

func (d *DB) IsFollower(claimer *claimer.Claimer, targetname *string) (isFollower bool, err error) {
	err = d.DB.Transaction(func(tx *gorm.DB) error {
		claimerId, err := getUserIdByClaimer(tx, claimer)
		if err != nil {
			return err
		}
		targetId, err := getUserIdByName(tx, targetname)
		if err != nil {
			return err
		}
		r := &model.UserFollow{}
		if err = tx.
			Select("").
			Where("user_id = ? AND target_id = ?", claimerId, targetId).
			First(r).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil
			}
			return err
		}
		isFollower = true
		return nil
	})
	return
}

func (d *DB) IsFollowing(claimer *claimer.Claimer, targetname *string) (isFollowing bool, err error) {
	err = d.DB.Transaction(func(tx *gorm.DB) error {
		claimerId, err := getUserIdByClaimer(tx, claimer)
		if err != nil {
			return err
		}
		targetId, err := getUserIdByName(tx, targetname)
		if err != nil {
			return err
		}
		r := &model.UserFollow{}
		if err = tx.
			Select("").
			Where("user_id = ? AND target_id = ?", targetId, claimerId).
			First(r).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil
			}
			return err
		}
		isFollowing = true
		return nil
	})
	return
}

func (d *DB) GetFollowers(claimer *claimer.Claimer, username *string, offset int, limit int) (*[]string, error) {
	if offset < 0 || limit < 1 || limit > 64 {
		return nil, ErrInvalidInput
	}

	followers := []string{}
	err := d.DB.Transaction(func(tx *gorm.DB) error {
		claimerId, err := getUserIdByClaimer(tx, claimer)
		if err != nil {
			return err
		}

		userId, err := getUserIdByName(tx, username)
		if err != nil {
			return err
		}

		ok, err := hasQueryPermission(tx, claimerId, userId)
		if err != nil {
			return err
		}

		if !ok {
			return ErrInvalidPermission
		}

		if err := tx.
			Model(&model.User{}).
			Select("users.username").
			Joins("JOIN user_follows ON user_follows.target_id = ? AND user_follows.user_id = users.id", userId).
			Offset(offset).
			Limit(limit).
			Find(&followers).Error; err != nil {
			return err
		}
		return nil
	})
	return &followers, err
}

func (d *DB) GetFollowings(claimer *claimer.Claimer, username *string, offset int, limit int) (*[]string, error) {
	if offset < 0 || limit < 1 || limit > 64 {
		return nil, ErrInvalidInput
	}

	followings := []string{}
	err := d.DB.Transaction(func(tx *gorm.DB) error {
		claimerId, err := getUserIdByClaimer(tx, claimer)
		if err != nil {
			return err
		}

		userId, err := getUserIdByName(tx, username)
		if err != nil {
			return err
		}

		ok, err := hasQueryPermission(tx, claimerId, userId)
		if err != nil {
			return err
		}

		if !ok {
			return ErrInvalidPermission
		}

		if err := tx.
			Model(&model.User{}).
			Select("users.username").
			Joins("JOIN user_follows ON user_follows.user_id = ? AND user_follows.target_id = users.id", userId).
			Offset(offset).
			Limit(limit).
			Find(&followings).Error; err != nil {
			return err
		}
		return nil
	})
	return &followings, err
}

func exchangeRequestCode(tx *gorm.DB, code *binaryuuid.UUID) (*model.UserFollowRequest, error) {
	request := &model.UserFollowRequest{}
	err := tx.
		Where("code = ?", code).
		First(&request).Error
	return request, err
}

func (d *DB) AcceptRequestFollow(claimer *claimer.Claimer, code *binaryuuid.UUID) error {
	return d.DB.Transaction(func(tx *gorm.DB) error {
		claimerId, err := getUserIdByClaimer(tx, claimer)
		if err != nil {
			return err
		}

		requestCode, err := exchangeRequestCode(tx, code)
		if err != nil {
			return err
		}
		if claimerId != requestCode.TargetId {
			return ErrInvalidInput
		}

		userId := requestCode.UserId

		followRequest := &model.UserFollowRequest{}
		if err := tx.
			Select("id", "user_id", "target_id").
			Where("user_id = ? AND target_id = ?", userId, claimerId).
			First(followRequest).Error; err != nil {
			return err
		}

		if err := tx.
			Delete(&model.UserFollowRequest{}, followRequest.Id).Error; err != nil {
			return err
		}

		err = followUser(tx, followRequest.UserId, followRequest.TargetId)
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return nil
		}
		return err
	})
}

func (d *DB) RejectRequestFollow(claimer *claimer.Claimer, code *binaryuuid.UUID) error {
	return d.DB.Transaction(func(tx *gorm.DB) error {
		claimerId, err := getUserIdByClaimer(tx, claimer)
		if err != nil {
			return err
		}

		requestUserId, err := exchangeRequestCode(tx, code)
		if err != nil {
			return err
		}

		followRequest := &model.UserFollowRequest{}
		if err := tx.
			Select("id", "user_id", "target_id").
			Where("user_id = ? AND target_id = ?", requestUserId, claimerId).
			First(followRequest).Error; err != nil {
			return err
		}

		return tx.Delete(&model.UserFollowRequest{}, followRequest.Id).Error
	})
}

func (d *DB) GetFollowRequests(claimer *claimer.Claimer, offset int, limit int) (*[]model.FollowRequest, error) {
	if offset < 0 || limit < 1 || limit > 64 {
		return nil, ErrInvalidInput
	}

	requests := []model.FollowRequest{}
	err := d.DB.Transaction(func(tx *gorm.DB) error {
		claimerId, err := getUserIdByClaimer(tx, claimer)
		if err != nil {
			return err
		}

		return tx.
			Model(&model.UserFollowRequest{}).
			Select("user_follow_requests.code", "users.username").
			Joins("JOIN users ON users.id = user_follow_requests.user_id").
			Where("target_id = ?", claimerId).
			Find(&requests).Error
	})
	return &requests, err
}

func (d *DB) RemoveFollower(claimer *claimer.Claimer, targetname *string) error {
	return d.DB.Transaction(func(tx *gorm.DB) error {
		claimerId, err := getUserIdByClaimer(tx, claimer)
		if err != nil {
			return err
		}

		targetId, err := getUserIdByName(tx, targetname)
		if err != nil {
			return err
		}

		result := tx.
			Where("user_id = ? AND target_id = ?", targetId, claimerId).
			Delete(&model.UserFollow{})
		if result.RowsAffected < 1 {
			return ErrNoAffectedRow
		}
		return result.Error
	})
}

func (d *DB) RemoveFollowing(claimer *claimer.Claimer, targetname *string) error {
	return d.DB.Transaction(func(tx *gorm.DB) error {
		claimerId, err := getUserIdByClaimer(tx, claimer)
		if err != nil {
			return err
		}

		targetId, err := getUserIdByName(tx, targetname)
		if err != nil {
			return err
		}

		result := tx.
			Where("user_id = ? AND target_id = ?", claimerId, targetId).
			Delete(&model.UserFollow{})
		if result.RowsAffected < 1 {
			return ErrNoAffectedRow
		}
		return result.Error
	})
}
