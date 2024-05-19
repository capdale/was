package database

import (
	"errors"
	"time"

	"github.com/capdale/was/model"
	"github.com/capdale/was/types/binaryuuid"
	"github.com/capdale/was/types/claimer"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var (
	ErrNoUserExist      = errors.New("no user exists")
	ErrTicketExpired    = errors.New("ticket expired")
	ErrEmailAlreadyUsed = errors.New("email already used")
)

// func (d *DB) ExchangeIDs2Names(ids *[]uint64) (*[]string, error) {
// 	names := make([]string, len(*ids))
// 	result := d.DB.
// 		Model(&model.User{}).
// 		Select("username").
// 		Where("id = ?", *ids).
// 		Find(&names)
// 	if result.Error != nil {
// 		return nil, result.Error
// 	}
// 	if result.RowsAffected != int64(len(*ids)) {
// 		return nil, ErrInvalidInput
// 	}
// 	return &names, nil
// }

func (d *DB) GetUserByEmail(email string) (user *model.User, err error) {
	user = &model.User{}
	err = d.DB.
		Where("email = ?", &email).
		First(user).Error
	return
}

func (d *DB) CreateWithGithub(username string, email string) (*model.User, error) {
	user := &model.User{
		Username:    username,
		Email:       email,
		AccountType: model.AccountTypeGithub,
		SocialUser: &model.SocialUser{
			AccountType: model.AccountTypeGithub,
		},
		UserDisplayType: &model.UserDisplayType{
			IsPrivate: false,
		},
	}
	err := d.DB.Create(user).Error
	return user, err
}

func (d *DB) CreateWithKakao(username string, email string) (*model.User, error) {
	user := &model.User{
		Username:    username,
		Email:       email,
		AccountType: model.AccountTypeKakao,
		SocialUser: &model.SocialUser{
			AccountType: model.AccountTypeKakao,
		},
		UserDisplayType: &model.UserDisplayType{
			IsPrivate: false,
		},
	}
	err := d.DB.Create(user).Error
	return user, err
}

func getUserIdByName(tx *gorm.DB, username *string) (uint64, error) {
	if username == nil {
		return 0, nil
	}

	var userId uint64
	if err := tx.
		Model(&model.User{}).
		Select("id").
		Where("username = ?", username).
		First(&userId).Error; err != nil {
		return 0, err
	}
	return userId, nil
}

func (d *DB) IsEmailUsed(email string) (bool, error) {
	if err := d.DB.
		Where("email = ?", email).
		First(&model.User{}).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return false, nil
		}
		return true, err
	}
	return true, nil
}

func (d *DB) CreateTicketByEmail(email string) (*binaryuuid.UUID, error) {
	// no transaction needed, if email duplicated, create account fail
	emailUsed, err := d.IsEmailUsed(email)
	if err != nil {
		return nil, err
	}
	if emailUsed {
		return nil, ErrEmailAlreadyUsed
	}

	ticket := &model.Ticket{
		Email: email,
	}
	if err := d.DB.Create(ticket).Error; err != nil {
		return nil, err
	}
	return &ticket.UUID, nil
}

func (d *DB) RemoveTicket(ticketUUID *binaryuuid.UUID) error {
	return d.DB.
		Where("uuid = ?", ticketUUID).
		Delete(&model.Ticket{}).Error
}

func (d *DB) IsTicketAvailable(ticketUUID *binaryuuid.UUID) (bool, error) {
	_, err := d.GetTicket(ticketUUID)
	if err != nil {
		if err == gorm.ErrRecordNotFound || err == ErrTicketExpired {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (d *DB) GetTicket(ticketUUID *binaryuuid.UUID) (*model.Ticket, error) {
	ticket := &model.Ticket{}
	if err := d.DB.
		Where("uuid = ?", ticketUUID).
		First(ticket).Error; err != nil {
		return nil, err
	}

	if ticket.CreatedAt.Add(time.Minute * 10).Before(time.Now()) {
		go d.RemoveTicket(ticketUUID)
		return nil, ErrTicketExpired
	}
	return ticket, nil
}

func (d *DB) GetEmailByTicket(ticketUUID *binaryuuid.UUID) (string, error) {
	ticket, err := d.GetTicket(ticketUUID)
	if err != nil {
		return "", err
	}
	return ticket.Email, nil
}

func (d *DB) CreateOriginViaTicket(ticketUUID *binaryuuid.UUID, username string, password string) error {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	ticket, err := d.GetTicket(ticketUUID)
	if err != nil {
		return err
	}

	return d.DB.Transaction(func(tx *gorm.DB) error {
		result := tx.
			Where("uuid = ?", ticket.UUID).
			Delete(ticket)
		if result.Error != nil {
			return result.Error
		}

		if result.RowsAffected < 1 {
			return ErrNoAffectedRow
		}

		return tx.Create(&model.User{
			Username:    username,
			Email:       ticket.Email,
			AccountType: model.AccountTypeOrigin,
			OriginUser: &model.OriginUser{
				Hashed: hashed,
			},
			UserDisplayType: &model.UserDisplayType{
				IsPrivate: false,
			},
		}).Error
	})
}

type userClaimdNhashed struct {
	AuthUUID binaryuuid.UUID
	Hashed   []byte
}

func (d *DB) GetOriginUserClaim(username string, password string) (*claimer.Claimer, error) {
	user := &userClaimdNhashed{}
	if err := d.DB.
		Model(&model.User{}).
		Select("users.auth_uuid", "origin_users.hashed").
		Joins("INNER JOIN origin_users ON origin_users.id = users.id").
		Where("username = ? AND account_type = ?", username, model.AccountTypeOrigin).
		First(user).Error; err != nil {
		return nil, err
	}
	if err := bcrypt.CompareHashAndPassword(user.Hashed, []byte(password)); err != nil {
		return nil, err
	}
	claimer := claimer.New(&user.AuthUUID)
	return claimer, nil
}

const (
	userVisibilityPublic  = 0
	userVisibilityPrivate = 1
)

func (d *DB) UserVisibilityPublic() int {
	return userVisibilityPublic
}

func (d *DB) UserVisibilityPrivate() int {
	return userVisibilityPrivate
}

func (d *DB) ChangeVisibility(claimer *claimer.Claimer, visibilityType int) error {
	return d.DB.Transaction(func(tx *gorm.DB) error {
		claimerId, err := getUserIdByClaimer(tx, claimer)
		if err != nil {
			return err
		}

		if err := tx.
			Model(&model.UserDisplayType{}).
			Where("user_id = ?", claimerId).
			Update("is_private", visibilityType == userVisibilityPrivate).Error; err != nil {
			return err
		}
		return nil
	})
}
