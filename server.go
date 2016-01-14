package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/vivowares/octopus/Godeps/_workspace/src/github.com/zenazn/goji/bind"
	"github.com/vivowares/octopus/Godeps/_workspace/src/github.com/zenazn/goji/graceful"
	"github.com/vivowares/octopus/configs"
	"github.com/vivowares/octopus/connections"
	"github.com/vivowares/octopus/handlers"
	"github.com/vivowares/octopus/models"
	. "github.com/vivowares/octopus/utils"
	"io/ioutil"
	"os"
	"path"
	"runtime"
	"strconv"
	"time"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	Initialize()

	go func() {
		Logger.Info(fmt.Sprintf("Octopus started listenning to port %d", configs.Config.Service.HttpPort))
		graceful.Serve(
			bind.Socket(":"+strconv.Itoa(configs.Config.Service.HttpPort)),
			HttpRouter(),
		)
	}()

	go func() {
		Logger.Info(fmt.Sprintf("Connection Manager started listenning to port %d", configs.Config.Service.WsPort))
		graceful.Serve(
			bind.Socket(":"+strconv.Itoa(configs.Config.Service.WsPort)),
			WsRouter(),
		)
	}()

	graceful.HandleSignals()
	graceful.PreHook(func() {
		Logger.Info("Octopus received signal, gracefully stopping...")
	})

	graceful.PostHook(func() {
		connections.CM.Close()
		Logger.Info("Waiting for websockets to drain...")
		time.Sleep(3 * time.Second)
		Logger.Info("Connection Manager closed.")
	})
	graceful.PostHook(func() { models.CloseDB() })
	graceful.PostHook(func() { models.CloseIndexClient() })
	graceful.PostHook(func() {
		Logger.Info("Octopus stopped")
	})
	graceful.PostHook(func() { CloseLogger() })
	graceful.PostHook(func() { removePidFile() })

	createPidFile()

	graceful.Wait()
}

func Initialize() {
	configFile := flag.String("conf", "", "config file location")
	flag.Parse()
	if len(*configFile) == 0 {
		defaultConf := "/etc/octopus/octopus.yml"
		if _, err := os.Stat(defaultConf); os.IsNotExist(err) {
			pwd, err := os.Getwd()
			PanicIfErr(err)
			*configFile = path.Join(pwd, "configs", "octopus_development.yml")
		} else {
			*configFile = defaultConf
		}
	}
	PanicIfErr(configs.InitializeConfig(*configFile))

	InitialLogger()
	p, _ := json.Marshal(configs.Config)
	Logger.Debug(string(p))
	PanicIfErr(models.InitializeDB())
	PanicIfErr(models.InitializeIndexClient())
	PanicIfErr(connections.InitializeCM())
	handlers.InitWsUpgrader()
}

func createPidFile() error {
	pid := os.Getpid()
	return ioutil.WriteFile(configs.Config.Service.PidFile, []byte(strconv.Itoa(pid)), 0644)
}

func removePidFile() error {
	return os.Remove(configs.Config.Service.PidFile)
}
