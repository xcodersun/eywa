package main

import (
	"flag"
	"fmt"
	"github.com/vivowares/octopus/Godeps/_workspace/src/github.com/rs/cors"
	"github.com/vivowares/octopus/Godeps/_workspace/src/github.com/zenazn/goji/bind"
	"github.com/vivowares/octopus/Godeps/_workspace/src/github.com/zenazn/goji/graceful"
	"github.com/vivowares/octopus/Godeps/_workspace/src/github.com/zenazn/goji/web"
	"github.com/vivowares/octopus/Godeps/_workspace/src/github.com/zenazn/goji/web/middleware"
	"github.com/vivowares/octopus/configs"
	"github.com/vivowares/octopus/connections"
	"github.com/vivowares/octopus/handlers"
	"github.com/vivowares/octopus/middlewares"
	"github.com/vivowares/octopus/models"
	. "github.com/vivowares/octopus/utils"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"runtime"
	"strconv"
	"time"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

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

	PanicIfErr(InitialLogger())
	PanicIfErr(models.InitializeDB())
	PanicIfErr(models.InitializeIndexClient())
	PanicIfErr(connections.InitializeCM())
	handlers.InitWsUpgrader()

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

func WsRouter() http.Handler {
	wsRouter := web.New()
	wsRouter.Use(middleware.RequestID)
	wsRouter.Use(middlewares.AccessLogging)
	wsRouter.Use(middleware.Recoverer)
	wsRouter.Use(middleware.AutomaticOptions)
	wsRouter.Get("/heartbeat", handlers.HeartBeatWs)
	wsRouter.Get("/ws/channels/:channel_id/devices/:device_id", handlers.WsHandler)

	wsRouter.Compile()

	return wsRouter
}

func HttpRouter() http.Handler {
	httpRouter := web.New()
	httpRouter.Use(middleware.RequestID)
	httpRouter.Use(middlewares.AccessLogging)
	httpRouter.Use(middleware.Recoverer)
	httpRouter.Use(middleware.AutomaticOptions)
	c := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
	})
	httpRouter.Use(c.Handler)

	httpRouter.Get("/heartbeat", handlers.HeartBeatHttp)

	httpRouter.Get("/channels", handlers.ListChannels)
	httpRouter.Post("/channels", handlers.CreateChannel)
	httpRouter.Get("/channels/:id", handlers.GetChannel)
	httpRouter.Get("/channels/:id/tag_stats", handlers.GetChannelTagStats)
	httpRouter.Get("/channels/:id/index_stats", handlers.GetChannelIndexStats)
	httpRouter.Delete("/channels/:id", handlers.DeleteChannel)
	httpRouter.Put("/channels/:id", handlers.UpdateChannel)

	httpRouter.Get("/dashboards", handlers.ListDashboards)
	httpRouter.Post("/dashboards", handlers.CreateDashboard)
	httpRouter.Get("/dashboards/:id", handlers.GetDashboard)
	httpRouter.Delete("/dashboards/:id", handlers.DeleteDashboard)
	httpRouter.Put("/dashboards/:id", handlers.UpdateDashboard)

	httpRouter.Get("/connections/_count", handlers.ConnectionCounts)
	httpRouter.Get("/channels/:id/value", handlers.QueryValue)
	httpRouter.Get("/channels/:id/series", handlers.QuerySeries)
	httpRouter.Compile()

	return httpRouter
}

func createPidFile() error {
	pid := os.Getpid()
	return ioutil.WriteFile(configs.Config.Service.PidFile, []byte(strconv.Itoa(pid)), 0644)
}

func removePidFile() error {
	return os.Remove(configs.Config.Service.PidFile)
}
