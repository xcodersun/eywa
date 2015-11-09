package persistence

import (
// "log"

// "github.com/boltdb/bolt"
// "github.com/spf13/viper"
// . "github.com/vivowares/octopus/utils"
)

// var DB bolt.DB

// func InitializeDB() {
// 	DB, err := bolt.Open(viper.GetString("persistence.db_file"), 0600, nil)
// 	PanicIfErr(err)
// }

var DB DataStore

type DataStore interface {
	Save() error
}
