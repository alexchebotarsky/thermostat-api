package handler

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/alexchebotarsky/thermostat-api/metrics"
	"github.com/alexchebotarsky/thermostat-api/model/thermostat"
	"github.com/alexchebotarsky/thermostat-api/processor/event"
)

type CurrentStateManager interface {
	FetchCurrentState(ctx context.Context, deviceID string) (*thermostat.CurrentState, error)
	UpdateCurrentState(context.Context, *thermostat.CurrentState) (*thermostat.CurrentState, error)
}

func CurrentState(manager CurrentStateManager) event.Handler {
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

		lastState, err := manager.FetchCurrentState(ctx, state.DeviceID)
		if err != nil {
			// Failed to fetch last known state, ignore
		} else if state.Timestamp.Before(lastState.Timestamp) {
			return fmt.Errorf("current state is older than the last known state for device %s", state.DeviceID)
		}

		updatedState, err := manager.UpdateCurrentState(ctx, &state)
		if err != nil {
			return fmt.Errorf("error updating current state: %v", err)
		}

		metrics.SetThermostatOperatingState(updatedState.DeviceID, updatedState.OperatingState)
		metrics.SetThermostatCurrentTemperature(updatedState.DeviceID, updatedState.CurrentTemperature)
		if updatedState.CurrentHumidity != nil {
			metrics.SetThermostatCurrentHumidity(updatedState.DeviceID, *updatedState.CurrentHumidity)
		}

		return nil
	}
}
