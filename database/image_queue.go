package database

import (
	"fmt"

	"github.com/capdale/was/location"
	"github.com/capdale/was/model"
	"github.com/google/uuid"
)

func (d *Database) PutInImageQueue(userUUID *uuid.UUID, fileUUID *uuid.UUID, geoLocation *location.GeoLocation) error {
	imageQueue := &model.ImageQueue{
		UUID:       *fileUUID,
		UserUUID:   *userUUID,
		Longtitude: geoLocation.Altitude,
		Latitude:   geoLocation.Latitude,
		Altitude:   geoLocation.Altitude,
		Accuracy:   geoLocation.Accuracy,
	}
	return d.DB.Create(imageQueue).Error
}

func (d *Database) PopImageQueues(n int) (*[]model.ImageQueue, error) {
	queue := []model.ImageQueue{}
	err := d.DB.Limit(n).Order("created_at ASC").Find(&queue).Error
	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}
	if len(queue) == 0 {
		return &queue, nil
	}
	tx := d.DB.Delete(&queue)
	if tx.Error != nil {
		tx.Rollback()
		return nil, tx.Error
	}
	return &queue, nil
}

func (d *Database) RecoverImageQueue(index uint) error {
	return d.DB.Model(&model.ImageQueue{ID: 1}).Update("deleted_at", nil).Error
}

func (d *Database) DeleteImageQueues(index uint) error {
	return d.DB.Unscoped().Delete(&model.ImageQueue{}, index).Error
}
