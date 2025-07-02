package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/alexchebotarsky/thermofridge-api/metrics"
	"github.com/alexchebotarsky/thermofridge-api/model/thermofridge"
	"github.com/go-chi/chi/v5"
)

type TargetStateFetcher interface {
	FetchTargetState(ctx context.Context, deviceID string) (*thermofridge.TargetState, error)
}

func GetTargetState(fetcher TargetStateFetcher) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		deviceID := chi.URLParam(r, "deviceID")

		state, err := fetcher.FetchTargetState(r.Context(), deviceID)
		if err != nil {
			HandleError(w, fmt.Errorf("error fetching target state: %v", err), http.StatusInternalServerError, true)
			return
		}

		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		err = json.NewEncoder(w).Encode(state)
		handleWritingErr(err)
	}
}

type TargetStateUpdater interface {
	UpdateTargetState(context.Context, *thermofridge.TargetState) (*thermofridge.TargetState, error)
}

type TargetStatePublisher interface {
	PublishTargetState(context.Context, *thermofridge.TargetState) error
}

func UpdateTargetState(updater TargetStateUpdater, publisher TargetStatePublisher) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var state thermofridge.TargetState
		err := json.NewDecoder(r.Body).Decode(&state)
		if err != nil {
			HandleError(w, fmt.Errorf("error decoding target state: %v", err), http.StatusBadRequest, false)
			return
		}

		deviceID := chi.URLParam(r, "deviceID")
		state.DeviceID = deviceID

		err = state.Validate()
		if err != nil {
			HandleError(w, fmt.Errorf("error validating target state: %v", err), http.StatusBadRequest, false)
			return
		}

		updatedState, err := updater.UpdateTargetState(r.Context(), &state)
		if err != nil {
			HandleError(w, fmt.Errorf("error updating target state: %v", err), http.StatusInternalServerError, true)
			return
		}

		err = publisher.PublishTargetState(r.Context(), updatedState)
		if err != nil {
			HandleError(w, fmt.Errorf("error publishing target state: %v", err), http.StatusInternalServerError, true)
			return
		}

		if updatedState.Mode != nil {
			metrics.SetThermofridgeMode(deviceID, *updatedState.Mode)
		}

		if updatedState.TargetTemperature != nil {
			metrics.SetThermofridgeTargetTemperature(deviceID, *updatedState.TargetTemperature)
		}

		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		err = json.NewEncoder(w).Encode(updatedState)
		handleWritingErr(err)
	}
}
