package main

import (
	"github.com/gorilla/mux"
	"github.com/spf13/viper"
	"github.com/vivowares/octopus/connections"
	"github.com/vivowares/octopus/handlers"
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
	connections.InitializeCM()

	go graceful.Serve(
		bind.Socket(":"+viper.GetString("http_port")),
		HttpRouter(),
	)

	go func() {
		http.ListenAndServe(":"+viper.GetString("ws_port"), WsRouter())
	}()

	graceful.HandleSignals()
	graceful.PreHook(func() { log.Printf("Goji received signal, gracefully stopping") })
	graceful.PreHook(func() { connections.CM.Close() })
	graceful.PostHook(func() { connections.CM.Wait() })
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
