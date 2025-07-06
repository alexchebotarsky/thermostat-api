package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"

	"github.com/alexchebotarsky/thermostat-api/app"
	"github.com/alexchebotarsky/thermostat-api/env"
	"github.com/alexchebotarsky/thermostat-api/logger"
	"github.com/alexchebotarsky/thermostat-api/metrics"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer cancel()

	env, err := env.LoadConfig(ctx)
	if err != nil {
		slog.Error(fmt.Sprintf("Error loading env config: %v", err))
		os.Exit(1)
	}

	logger.Init(env.LogLevel, env.LogFormat)

	err = metrics.Init()
	if err != nil {
		slog.Error(fmt.Sprintf("Error initializing metrics: %v", err))
	}

	app, err := app.New(ctx, env)
	if err != nil {
		slog.Error(fmt.Sprintf("Error creating app: %v", err))
		os.Exit(1)
	}

	app.Launch(ctx)
}
