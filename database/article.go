package database

import (
	"errors"

	"github.com/capdale/was/model"
	"github.com/capdale/was/types/binaryuuid"
)

var (
	ErrInvalidInput      = errors.New("invalid input")
	ErrInvalidPermission = errors.New("invalid permission")
)

func (d *DB) IsCollectionOwned(userId uint64, collectionUUIDs *[]binaryuuid.UUID) error {
	var count int64
	querys := make([]uint64, len(*collectionUUIDs))
	for i := 0; i < len(*collectionUUIDs); i++ {
		querys[i] = userId
	}

	if err := d.DB.
		Model(&model.Collection{}).
		Where("user_id = ? AND uuid = ?", querys, *collectionUUIDs).
		Count(&count).Error; err != nil {
		return err
	}

	if int(count) != len(*collectionUUIDs) {
		return ErrInvalidPermission
	}
	return nil
}

func (d *DB) CreateNewArticle(userId uint64, title string, content string, collectionUUIDs *[]binaryuuid.UUID, imageUUIDs *[]binaryuuid.UUID, collectionOrder *[]uint8) error {
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

	return d.DB.Create(&model.Article{
		UserID:             userId,
		Title:              title,
		Content:            content,
		ArticleCollections: collections,
		ArticleImages:      &images,
	}).Error
}

func (d *DB) GetArticle(claimerId uint64, linkId binaryuuid.UUID) (*model.ArticleAPI, error) {
	hasPermission, err := d.hasPermissionArticle(claimerId, &linkId)
	if err != nil {
		return nil, err
	}
	if !hasPermission {
		return nil, ErrInvalidPermission
	}

	article := &model.ArticleAPI{}
	if err = d.DB.
		Model(&model.Article{}).
		Preload("ArticleCollections").
		Preload("ArticleImages").
		Where("link_uuid = ?", linkId).
		First(article).Error; err != nil {
		return nil, err
	}
	return article, err
}

func (d *DB) hasPermissionArticle(claimerId uint64, linkId *binaryuuid.UUID) (bool, error) {
	var ownerId uint64
	if err := d.DB.
		Model(&model.Article{}).
		Select("user_id").
		Where("link_uuid = ?", linkId).
		First(&ownerId).Error; err != nil {
		return false, err
	}
	return d.HasQueryPermission(claimerId, ownerId)
}

func (d *DB) GetArticleLinkIdsByUserId(userId uint64, offset int, limit int) (*[]binaryuuid.UUID, error) {
	if offset < 0 || limit < 1 || limit > 100 {
		return nil, ErrInvalidInput
	}
	articles := make([]model.Article, limit)
	if err := d.DB.
		Select("LinkID").
		Where("user_uuid = ?", userId).
		Find(&articles).Error; err != nil {
		return nil, err
	}
	links := make([]binaryuuid.UUID, len(articles))
	for i, article := range articles {
		links[i] = article.LinkUUID
	}
	return &links, nil
}

func (d *DB) GetArticlesByUserUUID(userUUID binaryuuid.UUID, offset int, limit int) (*[]model.ArticleAPI, error) {
	userId, err := d.GetUserIdByUUID(userUUID)
	if err != nil {
		return nil, err
	}
	articles := make([]model.ArticleAPI, limit)
	if err = d.DB.
		Model(&model.Article{}).
		Where("user_id = ?", userId).
		Offset(offset).
		Limit(limit).
		Find(&articles).Error; err != nil {
		return nil, err
	}
	return &articles, nil
}

func (d *DB) HasAccessPermissionArticleImage(claimerId uint64, articleImageUUID *binaryuuid.UUID) (bool, error) {
	imageOwner, err := d.getArticleOwnerIdByArticleImage(articleImageUUID)
	if err != nil {
		return false, err
	}
	return d.HasQueryPermission(claimerId, imageOwner)
}

func (d *DB) getArticleOwnerIdByArticleImage(articleImageUUID *binaryuuid.UUID) (uint64, error) {
	var userId uint64
	if err := d.DB.
		Model(&model.Article{}).
		Select("article.user_id").
		Joins("JOIN article_images ON articles.id = article_images.id").
		Where("article_images.image_uuid = ?", articleImageUUID).
		First(&userId).Error; err != nil {
		return 0, err
	}
	return userId, nil
}

func (d *DB) DeleteArticle(claimerUUID *binaryuuid.UUID, articleLinkId *binaryuuid.UUID) error {
	userId, err := d.GetUserIdByUUID(*claimerUUID)
	if err != nil {
		return err
	}

	result := d.DB.Where("user_id = ? AND link_uuid = ?", userId, articleLinkId).Delete(&model.Article{})
	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected < 1 {
		return ErrNoAffectedRow
	}

	return nil
}
