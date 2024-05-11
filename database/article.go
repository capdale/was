package database

import (
	"errors"

	"github.com/capdale/was/model"
	"github.com/capdale/was/types/binaryuuid"
	"github.com/capdale/was/types/claimer"
	"gorm.io/gorm"
)

var (
	ErrInvalidInput      = errors.New("invalid input")
	ErrInvalidPermission = errors.New("invalid permission")
)

func (d *DB) IsCollectionOwned(claimer *claimer.Claimer, collectionUUIDs *[]binaryuuid.UUID) error {
	if claimer == nil {
		return model.ErrAnonymousQuery
	}

	collectionLength := len(*collectionUUIDs)
	if collectionLength < 0 {
		return ErrInvalidInput
	}

	var count int64
	querys := make([]uint64, collectionLength)
	return d.DB.Transaction(func(tx *gorm.DB) error {
		userId, err := getUserIdByClaimer(tx, claimer)
		if err != nil {
			return err
		}
		for i := 0; i < len(*collectionUUIDs); i++ {
			querys[i] = userId
		}

		if err := d.DB.
			Model(&model.Collection{}).
			Where("user_id = ? AND uuid = ?", querys, *collectionUUIDs).
			Count(&count).Error; err != nil {
			return err
		}

		if int(count) != collectionLength {
			return ErrInvalidPermission
		}
		return nil
	})
}

func (d *DB) CreateNewArticle(claimerUUID *claimer.Claimer, title string, content string, collectionUUIDs *[]binaryuuid.UUID, imageUUIDs *[]binaryuuid.UUID, collectionOrder *[]uint8) error {
	collections := make([]*model.ArticleCollection, len(*collectionUUIDs))
	for i, cuid := range *collectionUUIDs {
		collections[i] = &model.ArticleCollection{CollectionUUID: cuid, Order: (*collectionOrder)[i]}
	}

	images := make([]*model.ArticleImage, len(*imageUUIDs))
	for i, imageUUID := range *imageUUIDs {
		images[i] = &model.ArticleImage{
			ImageUUID: imageUUID,
			Order:     uint8(i),
		}
	}
	return d.DB.Transaction(func(tx *gorm.DB) error {
		claimerId, err := getUserIdByClaimer(tx, claimerUUID)
		if err != nil {
			return err
		}
		return tx.Create(&model.Article{
			UserID:             claimerId,
			Title:              title,
			Content:            content,
			ArticleCollections: collections,
			ArticleImages:      &images,
		}).Error
	})
}

func (d *DB) GetArticle(claimer *claimer.Claimer, linkId binaryuuid.UUID) (*model.ArticleAPI, error) {
	article := &model.ArticleAPI{}
	err := d.DB.Transaction(func(tx *gorm.DB) error {
		ok, err := hasPermissionArticle(tx, claimer, &linkId)
		if err != nil {
			return err
		}
		if !ok {
			return ErrInvalidPermission
		}

		return d.DB.
			Model(&model.Article{}).
			Preload("ArticleCollections").
			Preload("ArticleImages").
			Where("link_uuid = ?", linkId).
			First(article).Error
	})
	return article, err
}

func hasPermissionArticle(tx *gorm.DB, claimer *claimer.Claimer, linkId *binaryuuid.UUID) (bool, error) {
	var ok bool = false
	err := tx.Transaction(func(tx *gorm.DB) error {
		var ownerId uint64
		var claimerId uint64
		var err error
		if err = tx.
			Model(&model.Article{}).
			Select("user_id").
			Where("link_uuid = ?", linkId).
			First(&ownerId).Error; err != nil {
			return err
		}

		claimerId, err = getUserIdByClaimer(tx, claimer)
		if err != nil {
			return err
		}
		ok, err = hasQueryPermission(tx, claimerId, ownerId)
		return err
	})
	return ok, err
}

func (d *DB) GetArticleLinkIdsByUserUUID(claimer *claimer.Claimer, userUUID *binaryuuid.UUID, offset int, limit int) (*[]binaryuuid.UUID, error) {
	if offset < 0 || limit < 1 || limit > 100 {
		return nil, ErrInvalidInput
	}

	articles := make([]model.Article, limit)
	if err := d.DB.Transaction(func(tx *gorm.DB) error {
		claimerId, err := getUserIdByClaimer(tx, claimer)
		if err != nil {
			return err
		}

		userId, err := getUserIdByUUID(tx, userUUID)
		if err != nil {
			return err
		}
		ok, err := hasQueryPermission(tx, claimerId, userId)
		if err != nil {
			return err
		}

		if !ok {
			return ErrInvalidPermission
		}

		return tx.
			Select("LinkID").
			Where("user_uuid = ?", userUUID).
			Find(&articles).Error

	}); err != nil {
		return nil, err
	}

	links := make([]binaryuuid.UUID, len(articles))
	for i, article := range articles {
		links[i] = article.LinkUUID
	}
	return &links, nil
}

func (d *DB) HasAccessPermissionArticleImage(claimer *claimer.Claimer, articleImageUUID *binaryuuid.UUID) (bool, error) {
	var ok bool = false
	err := d.DB.Transaction(func(tx *gorm.DB) error {
		var err error
		imageOwner, err := getArticleOwnerIdByArticleImage(tx, articleImageUUID)
		if err != nil {
			return err
		}

		claimerId, err := getUserIdByClaimer(tx, claimer)
		if err != nil {
			return err
		}

		ok, err = hasQueryPermission(tx, claimerId, imageOwner)
		return err
	})
	return ok, err
}

func getArticleOwnerIdByArticleImage(tx *gorm.DB, articleImageUUID *binaryuuid.UUID) (uint64, error) {
	var userId uint64
	if err := tx.
		Model(&model.Article{}).
		Select("article.user_id").
		Joins("JOIN article_images ON articles.id = article_images.id").
		Where("article_images.image_uuid = ?", articleImageUUID).
		First(&userId).Error; err != nil {
		return 0, err
	}
	return userId, nil
}

func (d *DB) DeleteArticle(claimer *claimer.Claimer, articleLinkId *binaryuuid.UUID) error {
	return d.DB.Transaction(func(tx *gorm.DB) error {
		userId, err := getUserIdByClaimer(tx, claimer)
		if err != nil {
			return err
		}

		result := tx.
			Where("user_id = ? AND link_uuid = ?", userId, articleLinkId).
			Delete(&model.Article{})
		if result.Error != nil {
			return result.Error
		}

		if result.RowsAffected < 1 {
			return ErrNoAffectedRow
		}
		return nil
	})
}
