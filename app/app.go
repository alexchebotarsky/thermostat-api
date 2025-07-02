package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/alexchebotarsky/thermofridge-api/client/database"
	"github.com/alexchebotarsky/thermofridge-api/client/pubsub"
	"github.com/alexchebotarsky/thermofridge-api/env"
	"github.com/alexchebotarsky/thermofridge-api/processor"
	"github.com/alexchebotarsky/thermofridge-api/server"
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
		Database: clients.Database,
		PubSub:   clients.PubSub,
	})
	services = append(services, s)

	p := processor.New(processor.Clients{
		PubSub:   clients.PubSub,
		Database: clients.Database,
	})
	services = append(services, p)

	return services, nil
}

type Clients struct {
	Database *database.Database
	PubSub   *pubsub.PubSub
}

func setupClients(ctx context.Context, env *env.Config) (*Clients, error) {
	var c Clients
	var err error

	c.Database, err = database.New(env.DatabaseFilename, map[string]string{
		database.ModeKey:              env.DefaultMode,
		database.TargetTemperatureKey: fmt.Sprintf("%d", env.DefaultTargetTemperature),
	})
	if err != nil {
		return nil, fmt.Errorf("error creating new database client: %v", err)
	}

	c.PubSub, err = pubsub.New(ctx, env.PubSubHost, env.PubSubPort, env.PubSubClientID, env.PubSubQoS)
	if err != nil {
		return nil, fmt.Errorf("error creating new pubsub client: %v", err)
	}

	return &c, nil
}

func (c *Clients) Close() error {
	var errs []error

	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	return nil
}
