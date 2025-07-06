package handler

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/alexchebotarsky/thermostat-api/metrics"
	"github.com/alexchebotarsky/thermostat-api/model/thermostat"
	"github.com/alexchebotarsky/thermostat-api/processor/event"
)

type CurrentStateUpdater interface {
	UpdateCurrentState(context.Context, *thermostat.CurrentState) (*thermostat.CurrentState, error)
}

func CurrentState(updater CurrentStateUpdater) event.Handler {
	return func(ctx context.Context, payload []byte) error {
		var state thermostat.CurrentState
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

		metrics.SetThermostatOperatingState(updatedState.DeviceID, updatedState.OperatingState)
		metrics.SetThermostatCurrentTemperature(updatedState.DeviceID, updatedState.CurrentTemperature)

		return nil
	}
}
