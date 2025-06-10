package middleware

import (
	"context"
	"time"

	"github.com/alexchebotarsky/thermofridge-api/metrics"
	"github.com/alexchebotarsky/thermofridge-api/processor/event"
)

func Metrics(eventName string, next event.Handler) event.Handler {
	return func(ctx context.Context, payload []byte) error {
		start := time.Now()
		err := next(ctx, payload)
		duration := time.Since(start)

		var status string
		if err != nil {
			status = "ERR"
		} else {
			status = "OK"
		}

		metrics.AddEventProcessed(eventName, status)
		metrics.ObserveEventDuration(eventName, duration)

		return err
	}
}
