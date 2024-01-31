package database

import (
	"github.com/capdale/was/model"
	"github.com/google/uuid"
)

func (d *DB) GetCollections(userUUID *uuid.UUID, offset int, limit int) (*[]uuid.UUID, error) {
	collections := []model.Collection{}
	err := d.DB.Select("uuid").Where("user_uuid = ?", userUUID).Offset(offset).Limit(limit).Find(&collections).Error
	uuids := make([]uuid.UUID, len(collections))
	for i, collection := range collections {
		uuids[i] = collection.UUID
	}
	return &uuids, err
}

func (d *DB) GetCollectionByUUID(collectionUUID *uuid.UUID) (collection *model.CollectionAPI, err error) {
	collection = &model.CollectionAPI{}
	err = d.DB.Model(&model.Collection{}).Where("uuid = ?", collectionUUID).Find(collection).Error
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
		Geolocation:     collection.Geolocation,
	}).Error
	return
}
