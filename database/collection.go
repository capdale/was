package database

import (
	"github.com/capdale/was/model"
	"github.com/google/uuid"
)

func (d *DB) GetCollections(userUUID *uuid.UUID, offset int, limit int) (*[]model.CollectionAPI, error) {
	collections := []model.CollectionAPI{}
	err := d.DB.Where("user_uuid = ?", userUUID.String()).Offset(offset).Limit(limit).Find(&collections).Error
	return &collections, err
}

func (d *DB) GetCollectionByUUID(collectionUUID *uuid.UUID) (collection *model.CollectionAPI, err error) {
	err = d.DB.Where("uuid = ?", collectionUUID).Find(collection).Error
	return
}

func (d *DB) CreateCollectionWithUserUUID(collection *model.CollectionAPI, userUUID *uuid.UUID) (collectionUUID *uuid.UUID, err error) {
	uuid, err := uuid.NewRandom()
	if err != nil {
		return
	}
	collectionUUID = &uuid
	err = d.DB.Create(&model.Collection{
		UserUUID:        *userUUID,
		UUID:            uuid,
		CollectionIndex: collection.CollectionIndex,
		Longtitude:      collection.Longtitude,
		Latitude:        collection.Latitude,
		Altitude:        collection.Altitude,
		Accuracy:        collection.Accuracy,
	}).Error
	return
}
