package main

import (
	"github.com/vivowares/octopus/Godeps/_workspace/src/github.com/rs/cors"
	"github.com/vivowares/octopus/Godeps/_workspace/src/github.com/zenazn/goji/web"
	"github.com/vivowares/octopus/Godeps/_workspace/src/github.com/zenazn/goji/web/middleware"
	"github.com/vivowares/octopus/handlers"
	"github.com/vivowares/octopus/middlewares"
	"net/http"
)

func WsRouter() http.Handler {
	wsRouter := web.New()
	wsRouter.Use(middleware.RealIP)
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
	httpRouter.Use(middleware.RealIP)
	httpRouter.Use(middleware.RequestID)
	httpRouter.Use(middlewares.AccessLogging)
	httpRouter.Use(middlewares.Authenticator)
	httpRouter.Use(middleware.Recoverer)
	httpRouter.Use(middleware.AutomaticOptions)
	c := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
	})
	httpRouter.Use(c.Handler)

	httpRouter.Get("/heartbeat", handlers.HeartBeatHttp)

	httpRouter.Get("/configs", handlers.GetConfig)
	httpRouter.Get("/configs/_reload", handlers.ReloadConfig)

	httpRouter.Get("/login", handlers.Login)

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
