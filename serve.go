package main

import (
	"fmt"
	"github.com/eywa/configs"
	"github.com/eywa/connections"
	. "github.com/eywa/loggers"
	"github.com/eywa/models"
	"github.com/eywa/pubsub"
	"github.com/zenazn/goji/graceful"
	"io/ioutil"
	"log"
	"os"
	"strconv"
)

func serve() {
	cert := configs.Config().Security.SSL.CertFile
	key := configs.Config().Security.SSL.KeyFile
	if len(cert) > 0 {
		if _, err := os.Stat(cert); os.IsNotExist(err) {
			log.Fatalln(fmt.Sprintf("cert file doesn't exist at: %s\n", cert))
		}
	}
	if len(key) > 0 {
		if _, err := os.Stat(key); os.IsNotExist(err) {
			log.Fatalln(fmt.Sprintf("key file doesn't exist at: %s\n", key))
		}
	}

	go func() {
		if len(cert) > 0 && len(key) > 0 {
			Logger.Info(fmt.Sprintf("Eywa started listening to port %d with SSL", configs.Config().Service.ApiPort))
			graceful.ListenAndServeTLS(
				":"+strconv.Itoa(configs.Config().Service.ApiPort),
				cert,
				key,
				HttpRouter(),
			)
		} else {
			Logger.Info(fmt.Sprintf("Eywa started listening to port %d", configs.Config().Service.ApiPort))
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
		Logger.Info("Eywa received signal, gracefully stopping...")
	})

	graceful.PreHook(func() {
		close(connections.HttpCloseChan)
	})

	graceful.PostHook(func() {
		Logger.Info("Waiting for websockets to drain...")
		connections.CloseCMs()
		Logger.Info("Connection Manager closed.")
	})
	graceful.PostHook(func() { models.CloseDB() })
	graceful.PostHook(func() { models.CloseIndexClient() })
	graceful.PostHook(func() {
		Logger.Info("Eywa stopped")
	})
	graceful.PostHook(func() { CloseLogger() })
	graceful.PostHook(func() { pubsub.Close() })
	graceful.PostHook(func() { removePidFile() })

	createPidFile()

	graceful.Wait()
}

func createPidFile() error {
	pid := os.Getpid()
	return ioutil.WriteFile(configs.Config().Service.PidFile, []byte(strconv.Itoa(pid)), 0644)
}

func removePidFile() error {
	return os.Remove(configs.Config().Service.PidFile)
}
