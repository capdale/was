package database

import (
	"errors"

	"github.com/capdale/was/model"
	"github.com/capdale/was/types/binaryuuid"
)

var ErrNoUserExist = errors.New("no user exists")

func (d *DB) GetUserByEmail(email string) (user *model.User, err error) {
	user = &model.User{}
	err = d.DB.Where("email = ?", &email).First(user).Error
	return
}

func (d *DB) CreateWithGithub(username string, email string) (*model.User, error) {
	user := &model.User{
		Username: username,
		Email:    email,
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
