package database

import (
	"github.com/capdale/was/model"
	"github.com/google/uuid"
)

func (d *Database) PutCollectionFromImageQueue(m *model.ImageQueue, index int64) error {
	collection := &model.Collection{
		UUID:            m.UUID,
		UserUUID:        m.UserUUID,
		Longtitude:      m.Altitude,
		Latitude:        m.Latitude,
		Altitude:        m.Altitude,
		Accuracy:        m.Accuracy,
		OriginAt:        m.CreatedAt,
		CollectionIndex: index,
	}
	return d.DB.Create(collection).Error
}

func (d *Database) GetCollectection(userUUID *uuid.UUID, offset int, limit int) (*[]model.Collection, error) {
	collections := []model.Collection{}
	err := d.DB.Where("user_uuid = ?", userUUID.String()).Offset(offset).Limit(limit).Find(&collections).Error
	return &collections, err
}
