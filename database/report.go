package database

import (
	"errors"

	"github.com/capdale/was/model"
	"github.com/capdale/was/types/binaryuuid"
	"github.com/capdale/was/types/claimer"
	"gorm.io/gorm"
)

var (
	ErrInvalidDetailType = errors.New("error invalid detail type")
)

func (d *DB) CreateReportUser(claimer *claimer.Claimer, targetname *string, detailType int, description string) error {
	return d.DB.Transaction(func(tx *gorm.DB) error {
		issuerId, err := getUserIdByClaimer(tx, claimer)
		if err != nil {
			return err
		}

		targetId, err := getUserIdByName(tx, targetname)
		if err != nil {
			return err
		}

		reportUser := &model.ReportUser{
			ReportModel: model.ReportModel{
				IssuerId:    issuerId,
				Description: description,
			},
			TargetUserId:     targetId,
			ReportDetailType: detailType,
		}
		return tx.Create(reportUser).Error
	})

}

func (d *DB) CreateReportArticle(claimer *claimer.Claimer, linkUUID *binaryuuid.UUID, detailType int, description string) error {
	return d.DB.Transaction(func(tx *gorm.DB) error {
		var articleId uint64
		if err := tx.
			Select("id").
			Where("link_uuid = ?", linkUUID).
			First(&articleId).Error; err != nil {
			return err
		}

		issuerId, err := getUserIdByClaimer(tx, claimer)
		if err != nil {
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
		return tx.Create(reportArticle).Error
	})
}

func (d *DB) CreateReportBug(claimer *claimer.Claimer, title string, description string) error {
	return d.DB.Transaction(func(tx *gorm.DB) error {
		issuerId, err := getUserIdByClaimer(tx, claimer)
		if err != nil {
			return err
		}

		reportBug := &model.ReportBug{
			ReportModel: model.ReportModel{
				IssuerId:    issuerId,
				Description: description,
			},
		}
		return tx.Create(reportBug).Error
	})
}

func (d *DB) CreateReportHelp(claimer *claimer.Claimer, title string, description string) error {
	return d.DB.Transaction(func(tx *gorm.DB) error {
		issuerId, err := getUserIdByClaimer(tx, claimer)
		if err != nil {
			return err
		}

		reportHelp := &model.ReportHelp{
			ReportModel: model.ReportModel{
				IssuerId:    issuerId,
				Description: description,
			},
		}
		return tx.Create(reportHelp).Error
	})
}

func (d *DB) CreateReportEtc(claimer *claimer.Claimer, title string, description string) error {
	return d.DB.Transaction(func(tx *gorm.DB) error {
		issuerId, err := getUserIdByClaimer(tx, claimer)
		if err != nil {
			return err
		}
		reportEtc := &model.ReportEtc{
			ReportModel: model.ReportModel{
				IssuerId:    issuerId,
				Description: description,
			},
		}
		return tx.Create(reportEtc).Error
	})
}
