package handler

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/alexchebotarsky/thermofridge-api/model/thermofridge"
	"github.com/alexchebotarsky/thermofridge-api/processor/event"
)

type CurrentStateUpdater interface {
	UpdateCurrentState(*thermofridge.CurrentState) (*thermofridge.CurrentState, error)
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

		_, err = updater.UpdateCurrentState(&state)
		if err != nil {
			return fmt.Errorf("error updating current state: %v", err)
		}

		return nil
	}
}
