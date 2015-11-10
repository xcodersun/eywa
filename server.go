package main

import (
	"github.com/gorilla/mux"
	"github.com/spf13/viper"
	"github.com/vivowares/octopus/connections"
	"github.com/vivowares/octopus/handlers"
	"github.com/vivowares/octopus/models"
	. "github.com/vivowares/octopus/utils"
	"github.com/zenazn/goji/bind"
	"github.com/zenazn/goji/graceful"
	"github.com/zenazn/goji/web"
	"github.com/zenazn/goji/web/middleware"
	"log"
	"net/http"
	"os"
	"path"
)

func main() {
	configure()
	err := models.InitializeMetaStore()
	PanicIfErr(err)
	err = connections.InitializeCM()
	PanicIfErr(err)

	go func() {
		log.Printf("Goji started listenning to port %s", viper.GetString("http_port"))
		graceful.Serve(
			bind.Socket(":"+viper.GetString("http_port")),
			HttpRouter(),
		)
	}()

	go func() {
		log.Printf("Connection Manager started listenning to port %s", viper.GetString("ws_port"))
		http.ListenAndServe(":"+viper.GetString("ws_port"), WsRouter())
	}()

	graceful.HandleSignals()
	graceful.PreHook(func() { log.Printf("Goji received signal, gracefully stopping") })
	graceful.PreHook(func() { connections.CM.Close() })
	graceful.PostHook(func() {
		connections.CM.Wait()
		log.Printf("Connection Manager closed")
	})
	graceful.PostHook(func() { models.CloseMetaStore() })
	graceful.PostHook(func() { log.Printf("Goji stopped") })
	graceful.Wait()
}

func configure() {
	viper.SetConfigName("octopus")
	viper.AddConfigPath("/etc/octopus/configs/")
	pwd, err := os.Getwd()
	PanicIfErr(err)
	viper.AddConfigPath(path.Join(pwd, "configs"))
	err = viper.ReadInConfig()
	PanicIfErr(err)
}

func WsRouter() http.Handler {
	wsRouter := mux.NewRouter()
	wsRouter.HandleFunc("/ws/{device_id}", handlers.WsHandler)
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
	httpRouter.Compile()

	return httpRouter
}
