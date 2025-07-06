package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/alexchebotarsky/thermostat-api/client/pubsub"
	"github.com/alexchebotarsky/thermostat-api/client/storage"
	"github.com/alexchebotarsky/thermostat-api/env"
	"github.com/alexchebotarsky/thermostat-api/processor"
	"github.com/alexchebotarsky/thermostat-api/server"
)

type App struct {
	Services []Service
	Clients  *Clients
}

func New(ctx context.Context, env *env.Config) (*App, error) {
	var app App
	var err error

	app.Clients, err = setupClients(ctx, env)
	if err != nil {
		return nil, fmt.Errorf("error setting up clients: %v", err)
	}

	app.Services, err = setupServices(env, app.Clients)
	if err != nil {
		return nil, fmt.Errorf("error setting up services: %v", err)
	}

	return &app, nil
}

func (app *App) Launch(ctx context.Context) {
	errc := make(chan error, 1)

	for _, service := range app.Services {
		go service.Start(ctx, errc)
	}

	select {
	case <-ctx.Done():
		slog.Debug("Context is cancelled")
	case err := <-errc:
		slog.Error(fmt.Sprintf("Critical service error: %v", err))
	}

	var errs []error

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	for _, service := range app.Services {
		err := service.Stop(ctx)
		if err != nil {
			errs = append(errs, fmt.Errorf("error stopping a service: %v", err))
		}
	}

	err := app.Clients.Close()
	if err != nil {
		errs = append(errs, fmt.Errorf("error closing app clients: %v", err))
	}

	if len(errs) > 0 {
		slog.Error(fmt.Sprintf("Error gracefully shutting down: %v", errors.Join(errs...)))
	} else {
		slog.Debug("App has been gracefully shut down")
	}
}

type Service interface {
	Start(context.Context, chan<- error)
	Stop(context.Context) error
}

func setupServices(env *env.Config, clients *Clients) ([]Service, error) {
	var services []Service

	s := server.New(env.Host, env.Port, server.Clients{
		Storage: clients.Storage,
		PubSub:  clients.PubSub,
	})
	services = append(services, s)

	p := processor.New(processor.Clients{
		PubSub:  clients.PubSub,
		Storage: clients.Storage,
	})
	services = append(services, p)

	return services, nil
}

type Clients struct {
	Storage *storage.Client
	PubSub  *pubsub.Client
}

func setupClients(ctx context.Context, env *env.Config) (*Clients, error) {
	var c Clients
	var err error

	c.Storage, err = storage.New(ctx, env.StoragePath, env.DefaultMode, env.DefaultTargetTemperature)
	if err != nil {
		return nil, fmt.Errorf("error creating new storage client: %v", err)
	}

	c.PubSub, err = pubsub.New(ctx, env.PubSubHost, env.PubSubPort, env.PubSubClientID, env.PubSubQoS)
	if err != nil {
		return nil, fmt.Errorf("error creating new pubsub client: %v", err)
	}

	return &c, nil
}

func (c *Clients) Close() error {
	var errs []error

	err := c.Storage.Close()
	if err != nil {
		errs = append(errs, fmt.Errorf("error closing storage client: %v", err))
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	return nil
}
