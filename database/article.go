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

func (d *DB) IsCollectionOwned(userId int64, collectionUUIDs *[]binaryuuid.UUID) error {
	var count int64
	querys := make([]int64, len(*collectionUUIDs))
	for i := 0; i < len(*collectionUUIDs); i++ {
		querys[i] = userId
	}

	if err := d.DB.Model(&model.Collection{}).Where("user_id = ? AND uuid = ?", querys, *collectionUUIDs).Count(&count).Error; err != nil {
		return err
	}

	if int(count) != len(*collectionUUIDs) {
		return ErrInvalidPermission
	}
	return nil
}

func (d *DB) CreateNewArticle(userId int64, title string, content string, collectionUUIDs *[]binaryuuid.UUID, imageUUIDs *[]binaryuuid.UUID, collectionOrder *[]uint8) error {
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

func (d *DB) GetArticle(userId int64, linkId binaryuuid.UUID) (*model.ArticleAPI, error) {
	article := &model.ArticleAPI{}
	err := d.DB.Model(&model.Article{}).Preload("ArticleCollections").Preload("ArticleImages").Where("user_id = ? AND link_id = ?", userId, linkId).First(article).Error
	if err != nil {
		return nil, err
	}
	return article, err
}

func (d *DB) GetArticleLinkIdsByUserId(userId int64, offset int, limit int) (*[]binaryuuid.UUID, error) {
	if offset < 0 || limit < 1 || limit > 100 {
		return nil, ErrInvalidInput
	}
	articles := make([]model.Article, limit)
	if err := d.DB.Select("LinkID").Where("user_id = ?", userId).Find(&articles).Error; err != nil {
		return nil, err
	}
	links := make([]binaryuuid.UUID, len(articles))
	for i, article := range articles {
		links[i] = article.LinkID
	}
	return &links, nil
}

func (d *DB) GetArticlesByUserUUID(userUUID binaryuuid.UUID, offset int, limit int) (*[]model.ArticleAPI, error) {
	userId, err := d.GetUserIdByUUID(userUUID)
	if err != nil {
		return nil, err
	}
	articles := make([]model.ArticleAPI, limit)
	if err = d.DB.Model(&model.Article{}).Where("user_id = ?", userId).Offset(offset).Limit(limit).Find(&articles).Error; err != nil {
		return nil, err
	}
	return &articles, nil
}
