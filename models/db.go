package models

import (
	"github.com/jinzhu/gorm"
	_ "github.com/mattn/go-sqlite3"
	"github.com/waterwheel"
	. "github.com/eywa/configs"
	. "github.com/eywa/loggers"
)

var DB *gorm.DB

// Initialize database helper
func InitializeDB() error {
	db, err := gorm.Open(Config().Database.DbType, Config().Database.DbFile)
	if err != nil {
		return err
	}
	db.LogMode(waterwheel.MapLevel(Config().Logging.Database.Level) == waterwheel.Debug)
	db.SetLogger(DBLogger)
	DB = &db

	return nil
}

func CloseDB() error {
	return DB.Close()
}
