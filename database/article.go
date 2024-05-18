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

		if err := tx.
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
			UserID:      claimerId,
			Title:       title,
			Content:     content,
			Collections: collections,
			Images:      &images,
		}).Error
	})
}

func (d *DB) GetArticle(claimer *claimer.Claimer, linkId binaryuuid.UUID) (*model.ArticleAPI, error) {
	article := &model.ArticleAPI{}
	err := d.DB.Transaction(func(tx *gorm.DB) error {
		claimerId, err := getUserIdByClaimer(tx, claimer)
		if err != nil {
			return err
		}

		articleOwner, err := getArticleOwner(tx, &linkId)
		if err != nil {
			return err
		}

		ok, err := hasQueryPermission(tx, claimerId, articleOwner.UserId)
		if err != nil {
			return err
		}

		if !ok {
			return ErrInvalidPermission
		}

		return tx.
			Model(&model.Article{}).
			Preload("Collections").
			Preload("Images").
			Preload("Meta").
			Where(articleOwner.Id).
			First(article).Error
	})
	return article, err
}

type ArticleOwner struct {
	Id     uint64
	UserId uint64
}

func getArticleOwner(tx *gorm.DB, linkId *binaryuuid.UUID) (*ArticleOwner, error) {
	owner := &ArticleOwner{}
	err := tx.
		Model(&model.Article{}).
		Select("id, user_id").
		Where("link_uuid = ?", linkId).
		First(&owner).Error
	return owner, err
}

func (d *DB) GetArticleLinkIdsByUsername(claimer *claimer.Claimer, username *string, offset int, limit int) (*[]*binaryuuid.UUID, error) {
	if offset < 0 || limit < 1 || limit > 64 {
		return nil, ErrInvalidInput
	}

	links := []*binaryuuid.UUID{}
	if err := d.DB.Transaction(func(tx *gorm.DB) error {
		claimerId, err := getUserIdByClaimer(tx, claimer)
		if err != nil {
			return err
		}

		userId, err := getUserIdByName(tx, username)
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
			Model(&model.Article{}).
			Select("link_uuid").
			Where("user_id = ?", userId).
			Find(&links).
			Offset(offset).
			Limit(limit).Error

	}); err != nil {
		return nil, err
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

func (d *DB) GetComments(claimer *claimer.Claimer, articleLinkId *binaryuuid.UUID, offset uint, limit uint) (*[]model.ArticleCommentAPI, error) {
	comments := []model.ArticleCommentAPI{}
	err := d.DB.Transaction(func(tx *gorm.DB) error {
		claimerId, err := getUserIdByClaimer(tx, claimer)
		if err != nil {
			return err
		}

		articleOwner, err := getArticleOwner(tx, articleLinkId)
		if err != nil {
			return err
		}

		ok, err := hasQueryPermission(tx, claimerId, articleOwner.Id)
		if err != nil {
			return err
		}

		if !ok {
			return ErrInvalidPermission
		}

		if err := tx.
			Model(&model.ArticleComment{}).
			Joins("JOIN users ON article_comments.user_id == users.id").
			Select("users.username, article_comments.comment").
			Where("article_id = ?", articleOwner.Id).
			Find(&comments).Error; err != nil {
			return err
		}
		return nil
	})
	return &comments, err
}

func (d *DB) Comment(claimer *claimer.Claimer, articleLinkId *binaryuuid.UUID, comment *string) error {
	return d.DB.Transaction(func(tx *gorm.DB) error {
		claimerId, err := getUserIdByClaimer(tx, claimer)
		if err != nil {
			return err
		}

		articleOwner, err := getArticleOwner(tx, articleLinkId)
		if err != nil {
			return err
		}

		ok, err := hasQueryPermission(tx, claimerId, articleOwner.Id)
		if err != nil {
			return err
		}

		if !ok {
			return ErrInvalidPermission
		}

		return tx.Create(&model.ArticleComment{
			ArticleId: articleOwner.Id,
			UserId:    claimerId,
			Comment:   *comment,
		}).Error
	})
}

func (d *DB) GetHeartState(claimer *claimer.Claimer, articleUUID *binaryuuid.UUID) (bool, error) {
	var state bool = false
	err := d.DB.Transaction(func(tx *gorm.DB) error {
		claimerId, err := getUserIdByClaimer(tx, claimer)
		if err != nil {
			return err
		}

		articleOwner, err := getArticleOwner(tx, articleUUID)
		if err != nil {
			return err
		}

		ok, err := hasQueryPermission(tx, claimerId, articleOwner.Id)
		if err != nil {
			return err
		}

		if !ok {
			return ErrInvalidPermission
		}
		return tx.
			Model(&model.ArticleHeart{}).
			Select("count(*)>0").
			Where("article_id = ? AND user_id = ?", articleOwner.Id, claimerId).
			Find(&state).
			Error
	})
	return state, err
}

func (d *DB) DoHeart(claimer *claimer.Claimer, articleUUID *binaryuuid.UUID, action int) error {
	if action != 0 && action != 1 {
		return ErrInvalidInput
	}
	return d.DB.Transaction(func(tx *gorm.DB) error {
		claimerId, err := getUserIdByClaimer(tx, claimer)
		if err != nil {
			return err
		}

		articleOwner, err := getArticleOwner(tx, articleUUID)
		if err != nil {
			return err
		}

		ok, err := hasQueryPermission(tx, claimerId, articleOwner.UserId)
		if err != nil {
			return err
		}

		if !ok {
			return ErrInvalidPermission
		}
		heart := &model.ArticleHeart{
			ArticleId: articleOwner.Id,
			UserId:    claimerId,
		}
		var addHeart int
		if action == 1 {
			err = tx.Create(heart).Error
			addHeart = 1
		}
		if action == 0 {
			err = tx.
				Where("article_id = ? AND user_id = ?", heart.ArticleId, heart.UserId).
				Delete(&model.ArticleHeart{}).Error
			addHeart = -1
		}
		if err != nil {
			return err
		}

		return tx.
			Model(&model.ArticleMeta{}).
			Where("article_id = ?", articleOwner.Id).
			Update("heart_count", gorm.Expr("heart_count + ?", addHeart)).Error
	})
}

func (d *DB) CountHeart(claimer *claimer.Claimer, articleId *binaryuuid.UUID) (uint64, error) {
	var count uint64
	err := d.DB.Transaction(func(tx *gorm.DB) error {
		claimerId, err := getUserIdByClaimer(tx, claimer)
		if err != nil {
			return err
		}

		articleOwner, err := getArticleOwner(tx, articleId)
		if err != nil {
			return err
		}

		ok, err := hasQueryPermission(tx, claimerId, articleOwner.Id)
		if err != nil {
			return err
		}

		if !ok {
			return ErrInvalidPermission
		}
		return tx.
			Model(&model.ArticleMeta{}).
			Select("heart_count").
			Where("article_id = ?", articleOwner.Id).
			Find(&count).Error
	})
	return count, err
}
