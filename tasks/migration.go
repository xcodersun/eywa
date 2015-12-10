package main

import (
	"github.com/vivowares/octopus/configs"
	. "github.com/vivowares/octopus/models"
	. "github.com/vivowares/octopus/utils"
	"os"
	"path"
)

func main() {
	pwd, err := os.Getwd()
	PanicIfErr(err)
	configPath := path.Join(path.Dir(pwd), "configs")
	PanicIfErr(configs.InitializeConfig(configPath))
	PanicIfErr(InitializeDB())

	DB.AutoMigrate(
		&Channel{},
	)
}
