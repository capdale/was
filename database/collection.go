package database

import (
	"github.com/capdale/was/model"
	"github.com/capdale/was/types/binaryuuid"
	"github.com/capdale/was/types/claimer"
	"gorm.io/gorm"
)

func (d *DB) GetUserCollectionUUIDs(targetUUID *binaryuuid.UUID, offset int, limit int) (*[]binaryuuid.UUID, error) {
	collections := []model.Collection{}
	if err := d.DB.Transaction(func(tx *gorm.DB) error {
		claimerId, err := getUserIdByUUID(tx, targetUUID)
		if err != nil {
			return err
		}
		return tx.
			Model(&model.Collection{}).
			Select("uuid").
			Where("user_id = ?", claimerId).
			Offset(offset).
			Limit(limit).
			Find(&collections).Error
	}); err != nil {
		return nil, err
	}

	uuids := make([]binaryuuid.UUID, len(collections))
	for i, collection := range collections {
		uuids[i] = binaryuuid.UUID(collection.UUID)
	}
	return &uuids, nil
}

func (d *DB) GetCollectionByUUID(claimer *claimer.Claimer, collectionUUID *binaryuuid.UUID) (collection *model.CollectionAPI, err error) {
	d.DB.Transaction(func(tx *gorm.DB) error {
		var err error
		claimerId, err := getUserIdByClaimer(tx, claimer)
		if err != nil {
			return err
		}

		collection = &model.CollectionAPI{}
		if err = tx.
			Model(&model.Collection{}).
			Where("uuid = ?", collectionUUID).
			First(collection).Error; err != nil {
			return err
		}

		if collection.UserId == nil {
			return ErrNoAffectedRow
		}

		allowed, err := hasQueryPermission(tx, claimerId, *collection.UserId)
		if err != nil {
			return err
		}

		if !allowed {
			return ErrInvalidPermission
		}
		return nil
	})

	return
}

func (d *DB) CreateCollection(claimer *claimer.Claimer, collection *model.CollectionAPI, collectionUUID binaryuuid.UUID) error {
	return d.DB.Transaction(func(tx *gorm.DB) error {
		claimerId, err := getUserIdByClaimer(tx, claimer)
		if err != nil {
			return err
		}

		c := &model.Collection{
			UserId:          &claimerId,
			UUID:            collectionUUID,
			CollectionIndex: *collection.CollectionIndex,
			Geolocation:     collection.Geolocation,
		}
		return tx.Create(c).Error
	})
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
func (d *DB) HasAccessPermissionCollection(claimer *claimer.Claimer, collectionUUID binaryuuid.UUID) error {
	return d.DB.Transaction(func(tx *gorm.DB) error {
		var ownerId uint64
		if err := tx.
			Model(&model.Collection{}).
			Select("user_id").
			Where("uuid = ?", collectionUUID).
			First(&ownerId).Error; err != nil {
			return err
		}

		// claimer UUID has a high probability of query success being guaranteed, so query after collection query
		claimerId, err := getUserIdByClaimer(tx, claimer)
		if err != nil {
			return err
		}

		allowed, err := hasQueryPermission(tx, claimerId, ownerId)
		if err != nil {
			return err
		}

		if !allowed {
			return ErrInvalidPermission
		}
		return nil
	})
}

func (d *DB) DeleteCollection(claimer *claimer.Claimer, collectionUUID *binaryuuid.UUID) error {
	return d.DB.Transaction(func(tx *gorm.DB) error {
		userId, err := getUserIdByClaimer(tx, claimer)
		if err != nil {
			return err
		}

		result := tx.
			Where("user_id = ? AND uuid = ?", userId, collectionUUID).
			Delete(&model.Collection{})
		if result.Error != nil {
			return result.Error
		}

		if result.RowsAffected < 1 {
			return ErrNoAffectedRow
		}
		return nil
	})
}
