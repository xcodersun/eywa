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
			home := os.Getenv("OCTOPUS_HOME")
			if len(home) == 0 {
				panic("ENV OCTOPUS_HOME is not set")
			}

			*configFile = path.Join(home, "configs", "octopus_development.yml")
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
