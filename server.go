package main

import (
	"flag"
	"github.com/gorilla/mux"
	"github.com/vivowares/octopus/configs"
	"github.com/vivowares/octopus/connections"
	"github.com/vivowares/octopus/handlers"
	"github.com/vivowares/octopus/models"
	. "github.com/vivowares/octopus/utils"
	"github.com/zenazn/goji/bind"
	"github.com/zenazn/goji/graceful"
	"github.com/zenazn/goji/web"
	"github.com/zenazn/goji/web/middleware"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"strconv"
)

func main() {
	configFile := flag.String("conf", "", "config file location")
	flag.Parse()
	if len(*configFile) > 0 {
		PanicIfErr(configs.InitializeConfig(*configFile))
	} else {
		pwd, err := os.Getwd()
		PanicIfErr(err)
		*configFile = path.Join(pwd, "configs", "octopus_development.yml")
		PanicIfErr(configs.InitializeConfig(*configFile))
	}

	PanicIfErr(models.InitializeDB())
	PanicIfErr(models.InitializeIndexDB())
	PanicIfErr(connections.InitializeCM())

	go func() {
		log.Printf("Goji started listenning to port %s", configs.Config.Service.Host)
		graceful.Serve(
			bind.Socket(":"+strconv.Itoa(configs.Config.Service.HttpPort)),
			HttpRouter(),
		)
	}()

	go func() {
		log.Printf("Connection Manager started listenning to port %d", configs.Config.Service.WsPort)
		http.ListenAndServe(":"+strconv.Itoa(configs.Config.Service.WsPort), WsRouter())
	}()

	graceful.HandleSignals()
	graceful.PreHook(func() { log.Printf("Goji received signal, gracefully stopping") })
	graceful.PreHook(func() { connections.CM.Close() })

	graceful.PostHook(func() {
		connections.CM.Wait()
		log.Printf("Connection Manager closed")
	})
	graceful.PostHook(func() { models.CloseDB() })
	graceful.PostHook(func() { models.CloseIndexDB() })
	graceful.PostHook(func() { log.Printf("Goji stopped") })
	graceful.PostHook(func() { removePidFile() })

	createPidFile()

	graceful.Wait()
}

func WsRouter() http.Handler {
	wsRouter := mux.NewRouter()
	// wsRouter.HandleFunc("/ws/{channel_id}/{device_id}", handlers.WsHandler)
	wsRouter.HandleFunc("/heartbeat", handlers.HeartBeatWs)
	return wsRouter
}

func HttpRouter() http.Handler {
	httpRouter := web.New()
	httpRouter.Use(middleware.RequestID)
	httpRouter.Use(middleware.Logger)
	httpRouter.Use(middleware.Recoverer)
	httpRouter.Use(middleware.AutomaticOptions)
	httpRouter.Get("/heartbeat", handlers.HeartBeatHttp)

	httpRouter.Get("/channels", handlers.ListChannels)
	httpRouter.Post("/channels", handlers.CreateChannel)
	httpRouter.Get("/channels/:id", handlers.GetChannel)
	httpRouter.Delete("/channels/:id", handlers.DeleteChannel)
	httpRouter.Put("/channels/:id", handlers.UpdateChannel)

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
