package server

import (
	"github.com/alexchebotarsky/thermofridge-api/server/handler"
	"github.com/alexchebotarsky/thermofridge-api/server/middleware"
	chi "github.com/go-chi/chi/v5"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func (s *Server) setupRoutes() {
	s.Router.Get("/_healthz", handler.Health)
	s.Router.Handle("/metrics", promhttp.Handler())

	s.Router.Route(v1API, func(r chi.Router) {
		r.Use(middleware.Metrics)

		r.Get("/target-state", handler.GetTargetState(s.Clients.Database))
		r.Post("/target-state", handler.UpdateTargetState(s.Clients.Database, s.Clients.PubSub))

		r.Get("/current-state", handler.GetCurrentState(s.Clients.Database))
	})
}

const v1API = "/api/v1"
