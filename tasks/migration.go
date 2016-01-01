package main

import (
	"flag"
	"github.com/vivowares/octopus/configs"
	. "github.com/vivowares/octopus/models"
	. "github.com/vivowares/octopus/utils"
	"os"
	"path"
)

func main() {
	configFile := flag.String("conf", "", "config file location")
	flag.Parse()
	if len(*configFile) == 0 {
		defaultConf := "/etc/octopus/octopus.yml"
		if _, err := os.Stat(defaultConf); os.IsNotExist(err) {
			pwd, err := os.Getwd()
			PanicIfErr(err)
			*configFile = path.Join(path.Dir(pwd), "configs", "octopus_development.yml")
		} else {
			*configFile = defaultConf
		}
	}
	PanicIfErr(configs.InitializeConfig(*configFile))

	PanicIfErr(InitializeDB())

	DB.AutoMigrate(
		&Channel{},
		&Dashboard{},
	)
}
