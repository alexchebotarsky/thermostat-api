package server

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/alexchebotarsky/thermofridge-api/server/handler"
	chi "github.com/go-chi/chi/v5"
)

type Server struct {
	Host    string
	Port    uint16
	Router  chi.Router
	HTTP    *http.Server
	Clients Clients
}

type Clients struct {
	Database Database
	PubSub   PubSub
}

type Database interface {
	handler.TargetStateFetcher
	handler.TargetStateUpdater
	handler.CurrentStateFetcher
}

type PubSub interface {
	handler.TargetStatePublisher
}

func New(host string, port uint16, clients Clients) *Server {
	var s Server

	s.Host = host
	s.Port = port
	s.Router = chi.NewRouter()
	s.HTTP = &http.Server{
		Addr:         fmt.Sprintf("%s:%d", s.Host, s.Port),
		Handler:      s.Router,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}
	s.Clients = clients

	s.setupRoutes()

	return &s
}

func (s *Server) Start(ctx context.Context, errc chan<- error) {
	slog.Info(fmt.Sprintf("Server is listening at %s:%d", s.Host, s.Port))
	err := s.HTTP.ListenAndServe()
	if err != http.ErrServerClosed {
		errc <- fmt.Errorf("error listening and serving: %v", err)
	}
}

func (s *Server) Stop(ctx context.Context) error {
	err := s.HTTP.Shutdown(ctx)
	if err != nil {
		return fmt.Errorf("error shutting down http server: %v", err)
	}

	return nil
}
