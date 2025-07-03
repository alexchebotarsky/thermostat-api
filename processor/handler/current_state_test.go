package handler

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/alexchebotarsky/thermofridge-api/model/thermofridge"
)

type fakeCurrentStateUpdater struct {
	States map[string]thermofridge.CurrentState

	shouldFail bool
}

func (f *fakeCurrentStateUpdater) UpdateCurrentState(ctx context.Context, state *thermofridge.CurrentState) (*thermofridge.CurrentState, error) {
	if f.shouldFail {
		return nil, errors.New("test error")
	}

	if state != nil {
		f.States[state.DeviceID] = *state
	}

	return state, nil
}

func TestCurrentState(t *testing.T) {
	now := time.Now()

	type args struct {
		updater *fakeCurrentStateUpdater
		payload []byte
	}
	tests := []struct {
		name              string
		args              args
		wantErr           bool
		wantUpdaterStates map[string]thermofridge.CurrentState
	}{
		{
			name: "should update current state",
			args: args{
				updater: &fakeCurrentStateUpdater{
					States:     map[string]thermofridge.CurrentState{},
					shouldFail: false,
				},
				payload: []byte(
					fmt.Sprintf(`{
						"deviceId": "test_device_id",
						"operatingState": "HEATING",
						"currentTemperature": 18.8,
						"timestamp": "%s"
					}`, now.Add(-5*time.Minute).Format(time.RFC3339Nano)),
				),
			},
			wantErr: false,
			wantUpdaterStates: map[string]thermofridge.CurrentState{
				"test_device_id": {
					DeviceID:           "test_device_id",
					OperatingState:     thermofridge.HeatingOperatingState,
					CurrentTemperature: 18.8,
					Timestamp:          now.Add(-5 * time.Minute),
				},
			},
		},
		{
			name: "should error if payload is invalid JSON",
			args: args{
				updater: &fakeCurrentStateUpdater{
					States:     map[string]thermofridge.CurrentState{},
					shouldFail: false,
				},
				payload: []byte((`{
					"deviceId":
				}`)),
			},
			wantErr:           true,
			wantUpdaterStates: map[string]thermofridge.CurrentState{},
		},
		{
			name: "should error if payload has invalid values",
			args: args{
				updater: &fakeCurrentStateUpdater{
					States:     map[string]thermofridge.CurrentState{},
					shouldFail: false,
				},
				payload: []byte(
					fmt.Sprintf(`{
						"deviceId": "test_device_id",
						"operatingState": "HEATING",
						"currentTemperature": -100.0,
						"timestamp": "%s"
					}`, now.Add(-90*time.Minute).Format(time.RFC3339Nano)),
				),
			},
			wantErr:           true,
			wantUpdaterStates: map[string]thermofridge.CurrentState{},
		},
		{
			name: "should error if failed to update",
			args: args{
				updater: &fakeCurrentStateUpdater{
					States:     map[string]thermofridge.CurrentState{},
					shouldFail: true,
				},
				payload: []byte(
					fmt.Sprintf(`{
						"deviceId": "test_device_id",
						"operatingState": "HEATING",
						"currentTemperature": 18.8,
						"timestamp": "%s"
					}`, now.Add(-5*time.Minute).Format(time.RFC3339Nano)),
				),
			},
			wantErr:           true,
			wantUpdaterStates: map[string]thermofridge.CurrentState{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := CurrentState(tt.args.updater)
			err := handler(context.Background(), tt.args.payload)

			// Check expected error
			if (err != nil) != tt.wantErr {
				t.Errorf("CurrentState() error = %v, wantErr %v", err, tt.wantErr)
			}

			// Check updater states
			if len(tt.args.updater.States) != len(tt.wantUpdaterStates) {
				t.Errorf("CurrentState() len(updater.States) = %d, want %d", len(tt.args.updater.States), len(tt.wantUpdaterStates))
			}
			for deviceID, wantState := range tt.wantUpdaterStates {
				state, exists := tt.args.updater.States[deviceID]
				if !exists {
					t.Errorf("CurrentState() updater.States[%q] not found", deviceID)
					continue
				}

				if state.OperatingState != wantState.OperatingState {
					t.Errorf("CurrentState() updater.States[%q].OperationState = %v, want %v", deviceID, state.OperatingState, wantState.OperatingState)
				}

				if state.CurrentTemperature != wantState.CurrentTemperature {
					t.Errorf("CurrentState() updater.States[%q].TargetTemperature = %v, want %v", deviceID, state.CurrentTemperature, wantState.CurrentTemperature)
				}

				if !state.Timestamp.Equal(wantState.Timestamp) {
					t.Errorf("CurrentState() updater.States[%q].Timestamp = %v, want %v", deviceID, state.Timestamp, wantState.Timestamp)
				}
			}
		})
	}
}
