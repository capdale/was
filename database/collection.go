package database

import (
	"github.com/capdale/was/model"
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
