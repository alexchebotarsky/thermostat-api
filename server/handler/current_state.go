package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/alexchebotarsky/thermostat-api/client"
	"github.com/alexchebotarsky/thermostat-api/metrics"
	"github.com/alexchebotarsky/thermostat-api/model/thermostat"
	"github.com/go-chi/chi/v5"
)

type CurrentStateFetcher interface {
	FetchCurrentState(ctx context.Context, deviceID string) (*thermostat.CurrentState, error)
}

func GetCurrentState(fetcher CurrentStateFetcher) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		deviceID := chi.URLParam(r, "deviceID")

		state, err := fetcher.FetchCurrentState(r.Context(), deviceID)
		if err != nil {
			switch err.(type) {
			case *client.ErrNotFound:
				HandleError(w, fmt.Errorf("current state not found: %v", err), http.StatusNotFound, true)
			default:
				HandleError(w, fmt.Errorf("error fetching current state: %v", err), http.StatusInternalServerError, true)
			}
			return
		}

		err = state.Validate()
		if err != nil {
			HandleError(w, fmt.Errorf("error invalid current state: %v", err), http.StatusInternalServerError, true)
			return
		}

		metrics.SetThermostatOperatingState(deviceID, state.OperatingState)
		metrics.SetThermostatCurrentTemperature(deviceID, state.CurrentTemperature)

		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		err = json.NewEncoder(w).Encode(state)
		handleWritingErr(err)
	}
}
