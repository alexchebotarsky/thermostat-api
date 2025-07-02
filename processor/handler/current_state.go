package handler

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/alexchebotarsky/thermofridge-api/metrics"
	"github.com/alexchebotarsky/thermofridge-api/model/thermofridge"
	"github.com/alexchebotarsky/thermofridge-api/processor/event"
)

type CurrentStateUpdater interface {
	UpdateCurrentState(context.Context, *thermofridge.CurrentState) (*thermofridge.CurrentState, error)
}

func CurrentState(updater CurrentStateUpdater) event.Handler {
	return func(ctx context.Context, payload []byte) error {
		var state thermofridge.CurrentState
		err := json.Unmarshal(payload, &state)
		if err != nil {
			return fmt.Errorf("error unmarshalling current state: %v", err)
		}

		err = state.Validate()
		if err != nil {
			return fmt.Errorf("error validating current state: %v", err)
		}

		updatedState, err := updater.UpdateCurrentState(ctx, &state)
		if err != nil {
			return fmt.Errorf("error updating current state: %v", err)
		}

		metrics.SetThermofridgeOperatingState(updatedState.DeviceID, updatedState.OperatingState)
		metrics.SetThermofridgeCurrentTemperature(updatedState.DeviceID, updatedState.CurrentTemperature)

		return nil
	}
}
