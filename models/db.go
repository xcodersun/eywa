package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"github.com/vivowares/eywa/Godeps/_workspace/src/github.com/jinzhu/gorm"
	_ "github.com/vivowares/eywa/Godeps/_workspace/src/github.com/mattn/go-sqlite3"
	"github.com/vivowares/eywa/Godeps/_workspace/src/github.com/vivowares/waterwheel"
	. "github.com/vivowares/eywa/configs"
	. "github.com/vivowares/eywa/utils"
	"strings"
)

var DB *gorm.DB

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

type StringMap map[string]string

func (m *StringMap) Scan(value interface{}) error {
	return json.Unmarshal(value.([]byte), m)
}

func (m StringMap) Value() (driver.Value, error) {
	b, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}
	return string(b), nil
}

type StringSlice []string

func (s *StringSlice) Scan(src interface{}) error {
	asBytes, ok := src.([]byte)
	if !ok {
		return errors.New("scan source was not []bytes")
	}

	asString := string(asBytes)
	if len(asString) == 0 {
		(*s) = StringSlice([]string{})
	} else {
		parsed := strings.Split(asString, ",")
		(*s) = StringSlice(parsed)
	}

	return nil
}

func (s StringSlice) Value() (driver.Value, error) {
	return strings.Join(s, ","), nil
}
