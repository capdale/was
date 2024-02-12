package database

import (
	"errors"
	"fmt"

	"github.com/capdale/was/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

var ErrNoUserExist = errors.New("no user exists")

func (d *DB) GetUserByEmail(email string) (user *model.User, err error) {
	user = &model.User{}
	err = d.DB.Where("email = ?", &email).First(user).Error
	return
}

func (d *DB) IsUserUUIDExist(uuid *uuid.UUID) (err error) {
	user := new(model.User)
	err = d.DB.First(user, "uuid = ?", uuid.String()).Error
	if err == gorm.ErrRecordNotFound {
		return ErrNoUserExist
	}
	return
}

func (d *DB) CreateWithGithub(username string, email string) (*model.User, error) {
	uuid, err := uuid.NewRandom()
	if err != nil {
		return nil, err
	}
	user := &model.User{
		Username: username,
		UUID:     uuid,
		Email:    email,
	}
	fmt.Print(user)
	err = d.DB.Create(user).Error
	return user, err
}

func (d *DB) GetUsernameByUUID(uuid string) (*model.User, error) {
	user := &model.User{}
	err := d.DB.Select("username").First(user, "uuid = ?", uuid).Error
	return user, err
}

func (d *DB) GetUserIdByUUID(uuid *uuid.UUID) (int64, error) {
	user := new(model.User)
	if err := d.DB.Select("id").Where("uuid = ?", uuid).First(user).Error; err != nil {
		return 0, err
	}
	return user.Id, nil
}
