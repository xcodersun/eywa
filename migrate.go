package main

import (
	. "github.com/vivowares/eywa/models"
	. "github.com/vivowares/eywa/utils"
)

func migrate() {
	FatalIfErr(DB.AutoMigrate(
		&Channel{},
		&Dashboard{},
	).Error)
}
