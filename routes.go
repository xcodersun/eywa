package main

import (
	"github.com/vivowares/eywa/Godeps/_workspace/src/github.com/rs/cors"
	"github.com/vivowares/eywa/Godeps/_workspace/src/github.com/zenazn/goji/web"
	"github.com/vivowares/eywa/Godeps/_workspace/src/github.com/zenazn/goji/web/middleware"
	"github.com/vivowares/eywa/handlers"
	"github.com/vivowares/eywa/middlewares"
	"net/http"
)

func DeviceRouter() http.Handler {
	DeviceRouter := web.New()
	DeviceRouter.Use(middleware.RealIP)
	DeviceRouter.Use(middleware.RequestID)
	DeviceRouter.Use(middlewares.AccessLogging)
	DeviceRouter.Use(middleware.Recoverer)
	DeviceRouter.Use(middleware.AutomaticOptions)
	DeviceRouter.Get("/heartbeat", handlers.HeartBeatWs)
	DeviceRouter.Get("/ws/channels/:channel_id/devices/:device_id", handlers.WsHandler)
	DeviceRouter.Post("/channels/:channel_id/devices/:device_id/upload", handlers.HttpHandler)

	DeviceRouter.Compile()

	return DeviceRouter
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
		AllowedOrigins:   []string{"*"},
		AllowedHeaders:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "DELETE", "PUT"},
		AllowCredentials: true,
	})
	httpRouter.Use(c.Handler)

	//Admin Routes
	//Public Routes
	httpRouter.Get("/", handlers.Greeting)
	httpRouter.Get("/heartbeat", handlers.HeartBeatHttp)
	httpRouter.Get("/login", handlers.Login)

	//Protected Routes
	httpRouter.Get("/configs", handlers.GetConfig)
	httpRouter.Put("/configs", handlers.UpdateConfig)

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

	httpRouter.Get("/channels/:id/value", handlers.QueryValue)
	httpRouter.Get("/channels/:id/series", handlers.QuerySeries)
	httpRouter.Get("/channels/:id/raw", handlers.QueryRaw)

	httpRouter.Get("/ws/connections/count", handlers.ConnectionCounts)
	httpRouter.Get("/ws/channels/:channel_id/devices/:device_id/status", handlers.ConnectionStatus)

	httpRouter.Post("/channels/:channel_id/devices/:device_id/send", handlers.SendToDevice)
	httpRouter.Post("/channels/:channel_id/devices/:device_id/request", handlers.RequestToDevice)

	//API routes
	httpRouter.Get("/api/v1/channels/:id/value", handlers.QueryValue)
	httpRouter.Get("/api/v1/channels/:id/series", handlers.QuerySeries)
	httpRouter.Get("/api/v1/ws/channels/:channel_id/devices/:device_id/status", handlers.ConnectionStatus)
	httpRouter.Post("/api/v1/channels/:channel_id/devices/:device_id/send", handlers.SendToDevice)
	httpRouter.Post("/api/v1/channels/:channel_id/devices/:device_id/request", handlers.RequestToDevice)

	httpRouter.Compile()

	return httpRouter
}
