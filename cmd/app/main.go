package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/alexchebotarsky/thermofridge-api/app"
	"github.com/alexchebotarsky/thermofridge-api/env"
	"github.com/alexchebotarsky/thermofridge-api/logger"
	"github.com/alexchebotarsky/thermofridge-api/metrics"
)

func main() {
	ctx := context.Background()

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
