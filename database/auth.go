package database

import (
	"errors"
	"time"

	"github.com/capdale/was/model"
	"github.com/capdale/was/types/binaryuuid"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var ErrNoAffectedRow = errors.New("there is no specific row")

func (d *DB) SaveToken(issuerUUID binaryuuid.UUID, tokenString string, refreshToken *[]byte, agent *string) error {
	return d.DB.Create(&model.Token{
		IssuerUUID:   issuerUUID,
		Token:        tokenString,
		RefreshToken: *refreshToken,
		UserAgent:    *agent,
		ExpireAt:     time.Now().Add(time.Hour * 24),
	}).Error
}

func (d *DB) PopTokenByRefreshToken(refreshToken *[]byte, transactionF func(string) error) (tokenString *string, err error) {
	t := &model.Token{}
	d.DB.Transaction(func(tx *gorm.DB) error {
		if err = tx.Select("token").Where("refresh_token = ?", refreshToken).Find(t).Error; err != nil {
			return err
		}
		if err = tx.Select("token").Where("refresh_token = ?", refreshToken).Clauses(clause.Returning{}).Delete(t).Error; err != nil {
			return err
		}
		tokenString = &t.Token
		return transactionF(t.Token)
	})
	return
}

func (d *DB) IfTokenExistRemoveElseErr(tokenString string, until time.Duration, blackToken func(string, time.Duration) error) error {
	return d.DB.Transaction(func(tx *gorm.DB) error {
		result := tx.Select("").Where("token = ?", tokenString).Delete(&model.Token{})
		if result != nil {
			return tx.Error
		}
		if tx.RowsAffected < 1 {
			return ErrNoAffectedRow
		}

		if err := blackToken(tokenString, until); err != nil {
			return err
		}
		return nil
	})
}

func (d *DB) QueryAllTokensByUserUUID(userUUID *uuid.UUID) (*[]*model.Token, error) {
	tokenMs := []model.Token{}
	if err := d.DB.Where("user_uuid = ?", userUUID).Find(&tokenMs).Error; err != nil {
		return nil, err
	}

	deleteTokens := []string{}
	tokens := []*model.Token{}
	curTime := time.Now()

	for _, token := range tokenMs {
		if token.ExpireAt.Before(curTime) {
			tokens = append(tokens, &token)
		} else {
			deleteTokens = append(deleteTokens, token.Token)
		}
	}
	go d.RemoveTokens(&deleteTokens)
	return &tokens, nil
}

func (d *DB) RemoveTokens(tokenStrings *[]string) error {
	err := d.DB.Where("token = ?", tokenStrings).Delete(&model.Token{}).Error
	return err
}
