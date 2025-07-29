package handler

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/alexchebotarsky/thermostat-api/model/thermostat"
)

type fakeCurrentStateUpdater struct {
	States map[string]thermostat.CurrentState

	shouldFail bool
}

func (f *fakeCurrentStateUpdater) UpdateCurrentState(ctx context.Context, state *thermostat.CurrentState) (*thermostat.CurrentState, error) {
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
	initialCurrentHumidity := 43.3
	updatedCurrentHumidity := 45.6

	type args struct {
		updater *fakeCurrentStateUpdater
		payload []byte
	}
	tests := []struct {
		name              string
		args              args
		wantErr           bool
		wantUpdaterStates map[string]thermostat.CurrentState
	}{
		{
			name: "should update current state",
			args: args{
				updater: &fakeCurrentStateUpdater{
					States: map[string]thermostat.CurrentState{
						"test_device_id": {
							DeviceID:           "test_device_id",
							Timestamp:          now.Add(-30 * time.Minute),
							OperatingState:     thermostat.IdleOperatingState,
							CurrentTemperature: 17.5,
							CurrentHumidity:    &initialCurrentHumidity,
						},
					},
					shouldFail: false,
				},
				payload: []byte(
					fmt.Sprintf(`{
						"deviceId": "test_device_id",
						"timestamp": "%s",
						"operatingState": "HEATING",
						"currentTemperature": 18.8,
						"currentHumidity": 45.6
					}`, now.Add(-5*time.Minute).Format(time.RFC3339Nano)),
				),
			},
			wantErr: false,
			wantUpdaterStates: map[string]thermostat.CurrentState{
				"test_device_id": {
					DeviceID:           "test_device_id",
					Timestamp:          now.Add(-5 * time.Minute),
					OperatingState:     thermostat.HeatingOperatingState,
					CurrentTemperature: 18.8,
					CurrentHumidity:    &updatedCurrentHumidity,
				},
			},
		},
		{
			name: "should update current state without humidity",
			args: args{
				updater: &fakeCurrentStateUpdater{
					States: map[string]thermostat.CurrentState{
						"test_device_id": {
							DeviceID:           "test_device_id",
							Timestamp:          now.Add(-30 * time.Minute),
							OperatingState:     thermostat.IdleOperatingState,
							CurrentTemperature: 17.5,
						},
					},
					shouldFail: false,
				},
				payload: []byte(
					fmt.Sprintf(`{
						"deviceId": "test_device_id",
						"timestamp": "%s",
						"operatingState": "HEATING",
						"currentTemperature": 18.8
					}`, now.Add(-5*time.Minute).Format(time.RFC3339Nano)),
				),
			},
			wantErr: false,
			wantUpdaterStates: map[string]thermostat.CurrentState{
				"test_device_id": {
					DeviceID:           "test_device_id",
					Timestamp:          now.Add(-5 * time.Minute),
					OperatingState:     thermostat.HeatingOperatingState,
					CurrentTemperature: 18.8,
				},
			},
		},
		{
			name: "should error if payload is invalid JSON",
			args: args{
				updater: &fakeCurrentStateUpdater{
					States: map[string]thermostat.CurrentState{
						"test_device_id": {
							DeviceID:           "test_device_id",
							Timestamp:          now.Add(-30 * time.Minute),
							OperatingState:     thermostat.IdleOperatingState,
							CurrentTemperature: 17.5,
							CurrentHumidity:    &initialCurrentHumidity,
						},
					},
					shouldFail: false,
				},
				payload: []byte((`{
					"deviceId":
				}`)),
			},
			wantErr: true,
			wantUpdaterStates: map[string]thermostat.CurrentState{
				"test_device_id": {
					DeviceID:           "test_device_id",
					Timestamp:          now.Add(-30 * time.Minute),
					OperatingState:     thermostat.IdleOperatingState,
					CurrentTemperature: 17.5,
					CurrentHumidity:    &initialCurrentHumidity,
				},
			},
		},
		{
			name: "should error if payload has invalid values",
			args: args{
				updater: &fakeCurrentStateUpdater{
					States: map[string]thermostat.CurrentState{
						"test_device_id": {
							DeviceID:           "test_device_id",
							Timestamp:          now.Add(-30 * time.Minute),
							OperatingState:     thermostat.IdleOperatingState,
							CurrentTemperature: 17.5,
							CurrentHumidity:    &initialCurrentHumidity,
						},
					},
					shouldFail: false,
				},
				payload: []byte(
					fmt.Sprintf(`{
						"deviceId": "test_device_id",
						"timestamp": "%s",
						"operatingState": "HEATING",
						"currentTemperature": -100.0,
						"currentHumidity": 105.0
					}`, now.Add(-90*time.Minute).Format(time.RFC3339Nano)),
				),
			},
			wantErr: true,
			wantUpdaterStates: map[string]thermostat.CurrentState{
				"test_device_id": {
					DeviceID:           "test_device_id",
					Timestamp:          now.Add(-30 * time.Minute),
					OperatingState:     thermostat.IdleOperatingState,
					CurrentTemperature: 17.5,
					CurrentHumidity:    &initialCurrentHumidity,
				},
			},
		},
		{
			name: "should error if failed to update",
			args: args{
				updater: &fakeCurrentStateUpdater{
					States: map[string]thermostat.CurrentState{
						"test_device_id": {
							DeviceID:           "test_device_id",
							Timestamp:          now.Add(-30 * time.Minute),
							OperatingState:     thermostat.IdleOperatingState,
							CurrentTemperature: 17.5,
							CurrentHumidity:    &initialCurrentHumidity,
						},
					},
					shouldFail: true,
				},
				payload: []byte(
					fmt.Sprintf(`{
						"deviceId": "test_device_id",
						"timestamp": "%s",
						"operatingState": "HEATING",
						"currentTemperature": 18.8,
						"currentHumidity": 45.6
					}`, now.Add(-5*time.Minute).Format(time.RFC3339Nano)),
				),
			},
			wantErr: true,
			wantUpdaterStates: map[string]thermostat.CurrentState{
				"test_device_id": {
					DeviceID:           "test_device_id",
					Timestamp:          now.Add(-30 * time.Minute),
					OperatingState:     thermostat.IdleOperatingState,
					CurrentTemperature: 17.5,
					CurrentHumidity:    &initialCurrentHumidity,
				},
			},
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
					t.Errorf("CurrentState() updater.States[%s] not found", deviceID)
					continue
				}

				if !state.Timestamp.Equal(wantState.Timestamp) {
					t.Errorf("CurrentState() updater.States[%s].Timestamp = %v, want %v", deviceID, state.Timestamp, wantState.Timestamp)
				}

				if state.OperatingState != wantState.OperatingState {
					t.Errorf("CurrentState() updater.States[%s].OperationState = %v, want %v", deviceID, state.OperatingState, wantState.OperatingState)
				}

				if state.CurrentTemperature != wantState.CurrentTemperature {
					t.Errorf("CurrentState() updater.States[%s].TargetTemperature = %v, want %v", deviceID, state.CurrentTemperature, wantState.CurrentTemperature)
				}

				if !ptrEqual(state.CurrentHumidity, wantState.CurrentHumidity) {
					t.Errorf("CurrentState() updater.States[%s].CurrentHumidity = %v, want %v", deviceID, state.CurrentHumidity, wantState.CurrentHumidity)
				}
			}
		})
	}
}

func ptrEqual[T comparable](a, b *T) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}
