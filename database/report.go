package database

import (
	"errors"

	"github.com/capdale/was/model"
	"github.com/google/uuid"
)

var (
	ErrInvalidDetailType = errors.New("error invalid detail type")
)

func (d *DB) CreateReportUser(issuerId int64, targetUserUUID *uuid.UUID, detailType int, description string) error {
	reportUser := &model.ReportUser{
		ReportModel: model.ReportModel{
			IssuerId:    issuerId,
			Description: description,
		},
		TargetUserUUID:   *targetUserUUID,
		ReportDetailType: detailType,
	}
	return d.DB.Create(reportUser).Error
}

func (d *DB) CreateReportArticle(issuerId int64, targetArticleLink string, detailType int, description string) error {
	reportArticle := &model.ReportArticle{
		ReportModel: model.ReportModel{
			IssuerId:    issuerId,
			Description: description,
		},
		TargetArticleLink: targetArticleLink,
		ReportDetailType:  detailType,
	}
	return d.DB.Create(reportArticle).Error
}

func (d *DB) CreateReportBug(issuerId int64, title string, description string) error {
	reportBug := &model.ReportBug{
		ReportModel: model.ReportModel{
			IssuerId:    issuerId,
			Description: description,
		},
	}
	return d.DB.Create(reportBug).Error
}

func (d *DB) CreateReportHelp(issuerId int64, title string, description string) error {
	reportHelp := &model.ReportHelp{
		ReportModel: model.ReportModel{
			IssuerId:    issuerId,
			Description: description,
		},
	}
	return d.DB.Create(reportHelp).Error
}

func (d *DB) CreateReportEtc(issuerId int64, title string, description string) error {
	reportEtc := &model.ReportEtc{
		ReportModel: model.ReportModel{
			IssuerId:    issuerId,
			Description: description,
		},
	}
	return d.DB.Create(reportEtc).Error
}
