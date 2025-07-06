package server

import (
	"github.com/alexchebotarsky/thermostat-api/server/handler"
	"github.com/alexchebotarsky/thermostat-api/server/middleware"
	chi "github.com/go-chi/chi/v5"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func (s *Server) setupRoutes() {
	s.Router.Get("/_healthz", handler.Health)
	s.Router.Get("/openapi.yaml", handler.OpenapiYAML)
	s.Router.Get("/docs", handler.SwaggerUI)
	s.Router.Handle("/metrics", promhttp.Handler())

	s.Router.Route(v1API, func(r chi.Router) {
		r.Use(middleware.Metrics)

		r.Get("/target-state/{deviceID}", handler.GetTargetState(s.Clients.Storage))
		r.Post("/target-state/{deviceID}", handler.UpdateTargetState(s.Clients.Storage, s.Clients.PubSub))

		r.Get("/current-state/{deviceID}", handler.GetCurrentState(s.Clients.Storage))
	})
}

const v1API = "/api/v1"
