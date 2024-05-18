package database

import (
	"fmt"
	"time"

	"github.com/capdale/was/config"
	"github.com/capdale/was/model"
	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type DB struct {
	DB *gorm.DB
}

func (d *DB) Close() (err error) {
	sqlDB, err := d.DB.DB()
	if err != nil {
		return
	}
	err = sqlDB.Close()
	return
}

func New(databaseConfig *config.Database, lvl logger.LogLevel) (db *DB, err error) {
	if databaseConfig.SQLite != nil {
		return NewSQLite(databaseConfig.SQLite, lvl)
	}
	if databaseConfig.Mysql != nil {
		return NewMySQL(databaseConfig.Mysql, lvl)
	}
	return nil, config.ErrEmailConfig
}

func NewSQLite(sqliteConfig *config.SQLite, lvl logger.LogLevel) (db *DB, err error) {
	d, err := gorm.Open(sqlite.Open(sqliteConfig.Path), &gorm.Config{
		Logger: logger.Default.LogMode(lvl),
	})
	if err != nil {
		return
	}

	db = &DB{
		DB: d,
	}
	err = db.AutoMigrate()
	return
}

func NewMySQL(mysqlConfig *config.Mysql, lvl logger.LogLevel) (db *DB, err error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/was?charset=utf8mb4&parseTime=True&loc=Local",
		mysqlConfig.Username,
		mysqlConfig.Password,
		mysqlConfig.Address,
		mysqlConfig.Port,
	)

	d, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(lvl),
	})
	if err != nil {
		return
	}

	sqldb, err := d.DB()
	if err != nil {
		return
	}

	sqldb.SetMaxIdleConns(mysqlConfig.MaxIdleConns)
	sqldb.SetMaxOpenConns(mysqlConfig.MaxOpenConns)
	sqldb.SetConnMaxLifetime(time.Second * time.Duration(mysqlConfig.MaxLifetime))

	db = &DB{
		DB: d,
	}
	err = db.AutoMigrate()

	return
}

func (d *DB) AutoMigrate() (err error) {
	err = d.DB.AutoMigrate(
		&model.User{}, &model.Token{}, &model.SocialUser{}, &model.OriginUser{}, &model.Ticket{},
		&model.UserDisplayType{}, &model.UserFollow{}, &model.UserFollowRequest{},
		&model.Collection{},
		&model.ReportUser{}, &model.ReportArticle{}, &model.ReportBug{}, &model.ReportHelp{}, &model.ReportEtc{},
		&model.Article{}, &model.ArticleCollection{}, &model.ArticleImage{}, &model.ArticleMeta{}, &model.ArticleHeart{}, &model.ArticleComment{},
	)
	return err
}
