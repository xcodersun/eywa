package main

import (
	"flag"
	"github.com/vivowares/eywa/configs"
	. "github.com/vivowares/eywa/models"
	. "github.com/vivowares/eywa/utils"
	"os"
	"path"
)

func main() {
	configFile := flag.String("conf", "", "config file location")
	flag.Parse()
	if len(*configFile) == 0 {
		defaultConf := "/etc/eywa/eywa.yml"
		if _, err := os.Stat(defaultConf); os.IsNotExist(err) {
			home := os.Getenv("EYWA_HOME")
			if len(home) == 0 {
				panic("ENV EYWA_HOME is not set")
			}

			*configFile = path.Join(home, "configs", "eywa_development.yml")
		} else {
			*configFile = defaultConf
		}
	}
	PanicIfErr(configs.InitializeConfig(*configFile))
	InitialLogger()
	PanicIfErr(InitializeDB())

	DB.AutoMigrate(
		&Channel{},
		&Dashboard{},
	)
}
