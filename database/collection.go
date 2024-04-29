package database

import (
	"github.com/capdale/was/model"
	"github.com/capdale/was/types/binaryuuid"
)

func (d *DB) GetCollectionUUIDs(userId int64, offset int, limit int) (*[]binaryuuid.UUID, error) {
	collections := []model.Collection{}
	err := d.DB.
		Select("uuid").
		Where("user_id = ?", userId).
		Offset(offset).
		Limit(limit).
		Find(&collections).Error
	uuids := make([]binaryuuid.UUID, len(collections))
	for i, collection := range collections {
		uuids[i] = binaryuuid.UUID(collection.UUID)
	}
	return &uuids, err
}

func (d *DB) GetCollectionByUUID(claimerId int64, collectionUUID *binaryuuid.UUID) (collection *model.CollectionAPI, err error) {
	collection = &model.CollectionAPI{}
	if err = d.DB.
		Where("uuid = ?", collectionUUID).
		First(collection).Error; err != nil {
		return
	}

	allowed, err := d.HasQueryPermission(claimerId, collection.UserId)
	if err != nil {
		return
	}

	if !allowed {
		err = ErrInvalidPermission
		return
	}
	return
}

func (d *DB) CreateCollection(userId int64, collection *model.CollectionAPI, collectionUUID binaryuuid.UUID) error {
	c := &model.Collection{
		UserId:          userId,
		UUID:            collectionUUID,
		CollectionIndex: *collection.CollectionIndex,
		Geolocation:     collection.Geolocation,
	}
	err := d.DB.Create(c).Error
	return err
}

// func (d *DB) GetCollectionIdsByUUIDs(userId int64, collectionUUIDs *[]uuid.UUID) (*[]uint64, error) {
// 	query := []model.Collection{}
// 	for _, cuid := range *collectionUUIDs {
// 		query = append(query, model.Collection{UserId: userId, UUID: binaryuuid.UUID(cuid)})
// 	}
// 	collections := []model.Collection{}
// 	collectionIds := []uint64{}
// 	err := d.DB.Where(query, "userId", "UUID").Find(&collections).Error
// 	for _, collection := range collections {
// 		collectionIds = append(collectionIds, collection.Id)
// 	}
// 	return &collectionIds, err
// }

// TODO: if claimerUUID is nil, retrieve as public
// if collection is user owned return nil
func (d *DB) HasAccessPermissionCollection(claimerId int64, collectionUUID binaryuuid.UUID) error {
	var ownerId int64
	if err := d.DB.
		Select("user_id").
		Where("uuid = ?", collectionUUID).
		First(&ownerId).Error; err != nil {
		return err
	}

	allowed, err := d.HasQueryPermission(claimerId, ownerId)
	if err != nil {
		return err
	}
	if !allowed {
		return ErrInvalidPermission
	}
	return nil
}

func (d *DB) DeleteCollection(userUUID *binaryuuid.UUID, collectionUUID *binaryuuid.UUID) error {
	userId, err := d.GetUserIdByUUID(*userUUID)
	if err != nil {
		return err
	}

	result := d.DB.
		Where("user_id = ? AND uuid = ?", userId, collectionUUID).
		Delete(&model.Collection{})
	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected < 1 {
		return ErrNoAffectedRow
	}
	return nil
}
