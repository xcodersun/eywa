package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"github.com/eywa/configs"
	"github.com/eywa/connections"
	. "github.com/eywa/loggers"
	"github.com/eywa/models"
	. "github.com/eywa/utils"
	"os"
	"path"
	"runtime"
	"strings"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	args := os.Args
	tasks := []string{"serve", "migrate", "setup_es"}
	if len(args) < 2 {
		FatalIfErr(
			errors.New(fmt.Sprintf("task is not specified, available tasks are: %s", strings.Join(tasks, ","))),
		)
	}
	task := args[1]
	if !StringSliceContains(tasks, task) {
		FatalIfErr(
			errors.New(fmt.Sprintf("unknown task: %s, available tasks are: %s", task, strings.Join(tasks, ","))),
		)
	}

	initialize()

	switch args[1] {
	case "serve":
		FatalIfErr(models.InitializeDB())
		FatalIfErr(models.InitializeIndexClient())
		names := make([]string, 0)
		chs := models.Channels()
		for _, ch := range chs {
			name, err := ch.HashId()
			if err != nil {
				FatalIfErr(err)
			}
			names = append(names, name)
		}
		connections.InitWsUpgraders()
		FatalIfErr(connections.InitializeCMs(names))
		serve()
	case "migrate":
		FatalIfErr(models.InitializeDB())
		migrate()
	case "setup_es":
		FatalIfErr(models.InitializeIndexClient())
		setupES()
	}
}

func initialize() {
	fSet := flag.NewFlagSet("skip first arg", flag.ExitOnError)
	configFile := fSet.String("conf", "", "config file location")
	fSet.Parse(os.Args[2:])

	home := os.Getenv("EYWA_HOME")
	if len(home) == 0 {
		FatalIfErr(errors.New("ENV EYWA_HOME is not set"))
	}

	if len(*configFile) == 0 {
		defaultConf := "/etc/eywa/eywa.yml"
		if _, err := os.Stat(defaultConf); os.IsNotExist(err) {
			*configFile = path.Join(home, "configs", "eywa_development.yml")
		} else {
			*configFile = defaultConf
		}
	}

	params := map[string]string{"eywa_home": home}
	FatalIfErr(configs.InitializeConfig(*configFile, params))

	InitialLogger()
	p, _ := json.Marshal(configs.Config())
	Logger.Debug(string(p))
}
