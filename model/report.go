package model

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

//  don't use itoa

const (
	ReportTypeUser    = "report_user"
	ReportTypeArticle = "report_article"
	ReportTypeBug     = "report_bug"
	ReportTypeHelp    = "report_help"
	ReportTypeEtc     = "report_etc"
)

const (
	ReportUserMin       = 0
	ReportUserUsername  = 0
	ReportTypeUserAbuse = 1
	ReportUserMax       = 1
)

const (
	ReportArticleMin   = 0
	ReportArticleAbuse = 0
	ReportArticleEtc   = 1
	ReportArticleMax   = 1
)

type ReportModel struct {
	gorm.Model
	IssuerId    int64 `gorm:"index"`
	Description string
}

type ReportUser struct {
	ReportModel
	ReportDetailType int
	TargetUserUUID   uuid.UUID
}

type ReportArticle struct {
	ReportModel
	ReportDetailType  int
	TargetArticleLink string
}

type ReportBug struct {
	ReportModel
	Title string `gorm:"varchar(225)"`
}

type ReportHelp struct {
	ReportModel
	Title string `gorm:"varchar(225)"`
}

type ReportEtc struct {
	ReportModel
	Title string `gorm:"varchar(225)"`
}
