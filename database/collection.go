package database

import (
	"github.com/capdale/was/model"
	"github.com/capdale/was/types/binaryuuid"
	"github.com/google/uuid"
)

func (d *DB) GetCollectionUUIDs(userId int64, offset int, limit int) (*[]binaryuuid.UUID, error) {
	collections := []model.Collection{}
	err := d.DB.Select("uuid").Where("user_id = ?", userId).Offset(offset).Limit(limit).Find(&collections).Error
	uuids := make([]binaryuuid.UUID, len(collections))
	for i, collection := range collections {
		uuids[i] = binaryuuid.UUID(collection.UUID)
	}
	return &uuids, err
}

func (d *DB) GetCollectionByUUID(collectionUUID *uuid.UUID) (collection *model.CollectionAPI, err error) {
	collection = &model.CollectionAPI{}
	err = d.DB.Model(&model.Collection{}).Where("uuid = ?", collectionUUID).Find(collection).Error
	return
}

func (d *DB) CreateCollection(collection *model.CollectionAPI, userId int64) (binaryuuid.UUID, error) {
	c := &model.Collection{
		UserId:          userId,
		CollectionIndex: collection.CollectionIndex,
		Geolocation:     collection.Geolocation,
	}
	err := d.DB.Create(c).Error
	return c.UUID, err
}

func (d *DB) GetCollectionIdsByUUIDs(userId int64, collectionUUIDs *[]uuid.UUID) (*[]uint64, error) {
	query := []model.Collection{}
	for _, cuid := range *collectionUUIDs {
		query = append(query, model.Collection{UserId: userId, UUID: binaryuuid.UUID(cuid)})
	}
	collections := []model.Collection{}
	collectionIds := []uint64{}
	err := d.DB.Where(query, "userId", "UUID").Find(&collections).Error
	for _, collection := range collections {
		collectionIds = append(collectionIds, collection.Id)
	}
	return &collectionIds, err
}
