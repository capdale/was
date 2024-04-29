package database

import (
	"errors"

	"github.com/capdale/was/model"
	"github.com/capdale/was/types/binaryuuid"
)

var (
	ErrInvalidDetailType = errors.New("error invalid detail type")
)

func (d *DB) CreateReportUser(issuerId int64, targetUsername string, detailType int, description string) error {
	userId, err := d.GetUserIdByName(targetUsername)
	if err != nil {
		return err
	}

	reportUser := &model.ReportUser{
		ReportModel: model.ReportModel{
			IssuerId:    issuerId,
			Description: description,
		},
		TargetUserId:     userId,
		ReportDetailType: detailType,
	}
	return d.DB.Create(reportUser).Error
}

func (d *DB) CreateReportArticle(issuerId int64, linkUUID binaryuuid.UUID, detailType int, description string) error {
	var articleId uint64
	if err := d.DB.
		Select("id").
		Where("link_uuid = ?", linkUUID).
		First(&articleId).Error; err != nil {
		return err
	}
	reportArticle := &model.ReportArticle{
		ReportModel: model.ReportModel{
			IssuerId:    issuerId,
			Description: description,
		},
		TargetArticleId:  articleId,
		ReportDetailType: detailType,
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
