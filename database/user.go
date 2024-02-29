package database

import (
	"errors"
	"time"

	"github.com/capdale/was/model"
	"github.com/capdale/was/types/binaryuuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var (
	ErrNoUserExist      = errors.New("no user exists")
	ErrTicketExpired    = errors.New("ticket expired")
	ErrEmailAlreadyUsed = errors.New("email already used")
)

func (d *DB) GetUserById(userId int64) (*model.User, error) {
	user := &model.User{}
	err := d.DB.Where("id = ?", userId).First(user).Error
	return user, err
}

func (d *DB) GetUserByEmail(email string) (user *model.User, err error) {
	user = &model.User{}
	err = d.DB.Where("email = ?", &email).First(user).Error
	return
}

func (d *DB) CreateWithGithub(username string, email string) (*model.User, error) {
	user := &model.User{
		Username:    username,
		Email:       email,
		AccountType: model.AccountTypeGithub,
		SocialUser: model.SocialUser{
			AccountType: model.AccountTypeGithub,
		},
	}
	err := d.DB.Create(user).Error
	return user, err
}

func (d *DB) GetUserIdByUUID(userUUID binaryuuid.UUID) (int64, error) {
	user := new(model.User)
	if err := d.DB.Select("id").Where("uuid = ?", userUUID).First(user).Error; err != nil {
		return 0, err
	}
	return user.Id, nil
}

func (d *DB) IsEmailUsed(email string) (bool, error) {
	if err := d.DB.Where("email = ?", email).First(&model.User{}).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return false, nil
		}
		return true, err
	}
	return true, nil
}

func (d *DB) CreateTicketByEmail(email string) (*binaryuuid.UUID, error) {
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

func (d *DB) RemoveTicket(ticketUUID binaryuuid.UUID) error {
	return d.DB.Where("uuid = ?", ticketUUID).Delete(&model.Ticket{}).Error
}

func (d *DB) IsTicketAvailable(ticketUUID binaryuuid.UUID) (bool, error) {
	ticket, err := d.GetTicket(ticketUUID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return false, nil
		}
		return false, err
	}

	if ticket.CreatedAt.Add(time.Minute * 10).Before(time.Now()) {
		go d.RemoveTicket(ticketUUID)
		return false, nil
	}
	return true, nil
}

func (d *DB) GetTicket(ticketUUID binaryuuid.UUID) (*model.Ticket, error) {
	ticket := &model.Ticket{}
	if err := d.DB.Where("uuid = ?", ticketUUID).First(ticket).Error; err != nil {
		return nil, err
	}

	if ticket.CreatedAt.Add(time.Minute * 10).Before(time.Now()) {
		go d.RemoveTicket(ticketUUID)
		return nil, ErrTicketExpired
	}
	return ticket, nil
}

func (d *DB) CreateOriginViaTicket(ticketUUID binaryuuid.UUID, username string, password string) error {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	ticket, err := d.GetTicket(ticketUUID)
	if err != nil {
		return err
	}

	return d.DB.Create(&model.User{
		Username:    username,
		Email:       ticket.Email,
		AccountType: model.AccountTypeOrigin,
		OriginUser: model.OriginUser{
			Hashed: hashed,
		},
	}).Error
}

func (d *DB) GetOriginUserUUID(username string, password string) (*binaryuuid.UUID, error) {
	user := &model.User{}
	if err := d.DB.Preload("OriginUser", func(tx *gorm.DB) *gorm.DB {
		return tx.Select("hased")
	}).Select("uuid").Where("username = ? AND account_type = ?", username, model.AccountTypeOrigin).First(user).Error; err != nil {
		return nil, err
	}
	if err := bcrypt.CompareHashAndPassword(user.OriginUser.Hashed, []byte(password)); err != nil {
		return nil, err
	}
	return &user.UUID, nil
}
