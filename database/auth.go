package database

import (
	"errors"
	"time"

	"github.com/capdale/was/model"
	"github.com/capdale/was/types/binaryuuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var ErrNoAffectedRow = errors.New("there is no specific row")

func (d *DB) CreateRefreshToken(userId int64, refreshTokenUID *binaryuuid.UUID, refreshToken *[]byte, notBefore time.Time, expiredAt time.Time, agent *string) error {
	hashedToken, err := bcrypt.GenerateFromPassword(*refreshToken, bcrypt.MinCost)
	if err != nil {
		return err
	}

	return d.DB.Create(&model.Token{
		UserId:       userId,
		UUID:         *refreshTokenUID,
		RefreshToken: hashedToken,
		NotBefore:    notBefore,
		ExpireAt:     expiredAt,
		UserAgent:    *agent,
	}).Error
}

func (d *DB) PopRefreshToken(refreshTokenUID *binaryuuid.UUID) (*model.Token, error) {
	token := &model.Token{}
	err := d.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.
			Where("uuid = ?", refreshTokenUID).
			First(token).Error; err != nil {
			return err
		}
		if err := tx.
			Where("id = ?", token.Id).
			Delete(&model.Token{}).Error; err != nil {
			return err
		}
		return nil
	})
	return token, err

}

func (d *DB) RemoveRefreshToken(refreshToken *[]byte) error {
	result := d.DB.
		Where("refresh_token = ?", refreshToken).
		Delete(&model.Token{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected < 1 {
		return ErrNoAffectedRow
	}
	return nil
}

func (d *DB) QueryAllTokensByUserId(userId int64) (*[]*model.Token, error) {
	tokenMs := []model.Token{}
	if err := d.DB.
		Where("id = ?", userId).
		Find(&tokenMs).Error; err != nil {
		return nil, err
	}

	deleteRefreshTokens := []*[]byte{}
	tokens := []*model.Token{}
	curTime := time.Now()

	for _, token := range tokenMs {
		if token.ExpireAt.Before(curTime) {
			tokens = append(tokens, &token)
		} else {
			deleteRefreshTokens = append(deleteRefreshTokens, &token.RefreshToken)
		}
	}
	go d.RemoveTokens(&deleteRefreshTokens)
	return &tokens, nil
}

func (d *DB) RemoveTokens(refreshTokens *[]*[]byte) error {
	err := d.DB.
		Where("refresh_token = ?", refreshTokens).
		Delete(&model.Token{}).Error
	return err
}

func (d *DB) IsTokenPair(userId int64, tokenExpiredAt time.Time, refreshToken *[]byte) error {
	result := d.DB.
		Where("not_before = ? AND refresh_token = ? AND user_id = ?", tokenExpiredAt, refreshToken, userId).
		First(&model.Token{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected < 1 {
		return ErrNoAffectedRow
	}
	return nil
}

func (d *DB) DeleteUserAccount(claimerUUID *binaryuuid.UUID) error {
	return d.DB.Transaction(func(tx *gorm.DB) error {
		var claimerId int64
		if err := tx.
			Model(&model.User{}).
			Select("id").
			Where("uuid = ?", claimerUUID).
			First(&claimerId).Error; err != nil {
			return err
		}

		return tx.Delete(&model.User{}, claimerId).Error
	})
}
