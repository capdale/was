package database

import (
	"github.com/capdale/was/model"
)

func (d *Database) GetUserByEmail(email string) (*model.User, error) {
	user := &model.User{}
	err := d.DB.Where("email = ?", email).First(user).Error
	return user, err
}

func (d *Database) CreateSocial(user *model.User) error {
	return d.DB.Create(user).Error

}
