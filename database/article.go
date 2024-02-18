package database

import (
	"errors"

	"github.com/capdale/was/model"
	"github.com/capdale/was/types/binaryuuid"
)

var ErrInvalidInput = errors.New("invalid input")

func (d *DB) CreateNewArticle(userId int64, title string, content string, collectionUUIDs *[]binaryuuid.UUID) error {
	collections := []model.ArticleCollection{}

	for _, cuid := range *collectionUUIDs {
		collections = append(collections, model.ArticleCollection{CollectionUUID: cuid})
	}

	return d.DB.Create(&model.Article{
		UserID:             userId,
		Title:              title,
		Content:            content,
		ArticleCollections: collections,
	}).Error
}

func (d *DB) GetArticle(userId int64, linkId binaryuuid.UUID) (*model.ArticleAPI, error) {
	article := &model.ArticleAPI{}
	err := d.DB.Model(&model.Article{}).Where("user_id = ? AND link_id = ?", userId, linkId).First(article).Error
	if err != nil {
		return nil, err
	}

	articleCollections := &[]model.ArticleCollection{}
	if err = d.DB.Select("collection_uuid").Where("article_id = ?", article.Id).Find(articleCollections).Error; err != nil {
		return nil, err
	}

	articleCollectionUUIDs := make([]binaryuuid.UUID, len(*articleCollections))
	for i, articleCollection := range *articleCollections {
		articleCollectionUUIDs[i] = articleCollection.CollectionUUID
	}
	article.ArticleCollectionUUIDs = articleCollectionUUIDs
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
	if err = d.DB.Model(&model.Article{}).Preload("ArticleCollections").Where("user_id = ?", userId).Offset(offset).Limit(limit).Find(&articles).Error; err != nil {
		return nil, err
	}
	return &articles, nil
}
