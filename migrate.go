package main

import (
	. "github.com/eywa/models"
	. "github.com/eywa/utils"
)

func migrate() {
	FatalIfErr(DB.AutoMigrate(
		&Channel{},
		&Dashboard{},
	).Error)
}
