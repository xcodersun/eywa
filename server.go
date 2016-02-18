package main

import (
	"encoding/json"
	"flag"
	"fmt"
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

	cert := configs.Config().Security.SSL.CertFile
	key := configs.Config().Security.SSL.KeyFile
	if len(cert) > 0 {
		if _, err := os.Stat(cert); os.IsNotExist(err) {
			panic(fmt.Sprintf("cert file doesn't exist at: %s\n", cert))
		}
	}
	if len(key) > 0 {
		if _, err := os.Stat(key); os.IsNotExist(err) {
			panic(fmt.Sprintf("key file doesn't exist at: %s\n", key))
		}
	}

	go func() {
		if len(cert) > 0 && len(key) > 0 {
			Logger.Info(fmt.Sprintf("Octopus started listening to port %d with SSL", configs.Config().Service.ApiPort))
			graceful.ListenAndServeTLS(
				":"+strconv.Itoa(configs.Config().Service.ApiPort),
				cert,
				key,
				HttpRouter(),
			)
		} else {
			Logger.Info(fmt.Sprintf("Octopus started listening to port %d", configs.Config().Service.ApiPort))
			graceful.ListenAndServe(
				":"+strconv.Itoa(configs.Config().Service.ApiPort),
				HttpRouter(),
			)
		}
	}()

	go func() {
		if len(cert) > 0 && len(key) > 0 {
			Logger.Info(fmt.Sprintf("Connection Manager started listening to port %d with SSL", configs.Config().Service.DevicePort))
			graceful.ListenAndServeTLS(
				":"+strconv.Itoa(configs.Config().Service.DevicePort),
				cert,
				key,
				DeviceRouter(),
			)
		} else {
			Logger.Info(fmt.Sprintf("Connection Manager started listening to port %d", configs.Config().Service.DevicePort))
			graceful.ListenAndServe(
				":"+strconv.Itoa(configs.Config().Service.DevicePort),
				DeviceRouter(),
			)
		}

	}()

	graceful.HandleSignals()
	graceful.PreHook(func() {
		Logger.Info("Octopus received signal, gracefully stopping...")
	})

	graceful.PostHook(func() {
		connections.CloseWSCM()
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
	p, _ := json.Marshal(configs.Config())
	Logger.Debug(string(p))
	PanicIfErr(models.InitializeDB())
	PanicIfErr(models.InitializeIndexClient())
	PanicIfErr(connections.InitializeWSCM())
	handlers.InitWsUpgrader()
}

func createPidFile() error {
	pid := os.Getpid()
	return ioutil.WriteFile(configs.Config().Service.PidFile, []byte(strconv.Itoa(pid)), 0644)
}

func removePidFile() error {
	return os.Remove(configs.Config().Service.PidFile)
}
