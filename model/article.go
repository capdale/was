package model

import (
	"time"

	"github.com/capdale/was/types/binaryuuid"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Article struct {
	Id                 uint64          `gorm:"primaryKey"`
	UserID             uint64          `gorm:"index;index:uid_link_uuid_idx,unique;not null"` // = UserId
	User               User            `gorm:"references:id"`
	LinkUUID           binaryuuid.UUID `gorm:"index:uid_link_uuid_idx,unique;not null;"`
	Title              string          `gorm:"type:varchar(32);not null"`
	Content            string          `gorm:"type:TEXT;"`
	CreateAt           time.Time       `gorm:"autoCreateTime"`
	UpdateAt           time.Time       `gorm:"autoUpdateTime"`
	DeletedAt          gorm.DeletedAt
	Tags               []*ArticleTag
	ViewCount          uint64
	ArticleImages      *[]*ArticleImage     `gorm:"foreignKey:ArticleId;references:Id;contraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	ArticleCollections []*ArticleCollection `gorm:"foreignKey:ArticleId;references:Id;contraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}

func (a *Article) BeforeCreate(tx *gorm.DB) error {
	uid, err := uuid.NewRandom()
	a.LinkUUID = binaryuuid.UUID(uid)
	return err
}

type ArticleImage struct {
	ArticleId uint64          `gorm:"index" json:"-"`
	ImageUUID binaryuuid.UUID `gorm:"uuid;index" json:"uuid"`
	Order     uint8           `json:"order"`
}

type ArticleCollection struct {
	ArticleId      uint64          `gorm:"index" json:"-"`
	CollectionUUID binaryuuid.UUID `gorm:"index" json:"uuid"`
	Order          uint8           `json:"order"`
}

type ArticleTag struct {
	ArticleId uint64 `gorm:"index"`
	Tag       string `gorm:"index;varchar(12);not null"`
}

type ArticleComment struct {
	ArticleId uint64 `gorm:"index"`
	Comment   string `grom:"varchar(225);not null"`
}

type ArticleAPI struct {
	Id                 uint64               `json:"-"`
	Title              string               `json:"title"`
	Content            string               `json:"content"`
	UpdateAt           time.Time            `json:"update_at"`
	ArticleCollections []*ArticleCollection `json:"collections" gorm:"foreignKey:ArticleId;references:Id"`
	ArticleImages      []*ArticleImage      `json:"images" gorm:"foreignKey:ArticleId;references:Id"`
}
