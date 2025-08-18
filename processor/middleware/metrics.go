package middleware

import (
	"context"
	"encoding/json"
	"time"

	"github.com/alexchebotarsky/thermostat-api/metrics"
	"github.com/alexchebotarsky/thermostat-api/processor/event"
)

type DevicePayload struct {
	DeviceID string `json:"deviceId"`
}

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

		var devicePayload DevicePayload
		err = json.Unmarshal(payload, &devicePayload)
		if err != nil {
			devicePayload.DeviceID = "n/a"
		}

		metrics.AddEventProcessed(eventName, status, devicePayload.DeviceID)
		metrics.ObserveEventDuration(eventName, status, devicePayload.DeviceID, duration)

		return err
	}
}
