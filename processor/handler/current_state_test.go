package handler

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/alexchebotarsky/thermostat-api/model/thermostat"
)

type fakeCurrentStateManager struct {
	States map[string]thermostat.CurrentState

	shouldFail bool
}

func (f *fakeCurrentStateManager) FetchCurrentState(ctx context.Context, deviceID string) (*thermostat.CurrentState, error) {
	if f.shouldFail {
		return nil, errors.New("test error")
	}

	state, exists := f.States[deviceID]
	if !exists {
		return nil, fmt.Errorf("state not found for device %s", deviceID)
	}

	return &state, nil
}

func (f *fakeCurrentStateManager) UpdateCurrentState(ctx context.Context, state *thermostat.CurrentState) (*thermostat.CurrentState, error) {
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
		manager *fakeCurrentStateManager
		payload []byte
	}
	tests := []struct {
		name              string
		args              args
		wantErr           bool
		wantManagerStates map[string]thermostat.CurrentState
	}{
		{
			name: "should update current state",
			args: args{
				manager: &fakeCurrentStateManager{
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
						"operatingState": "COOLING",
						"currentTemperature": 18.8,
						"currentHumidity": 45.6
					}`, now.Add(-5*time.Minute).Format(time.RFC3339Nano)),
				),
			},
			wantErr: false,
			wantManagerStates: map[string]thermostat.CurrentState{
				"test_device_id": {
					DeviceID:           "test_device_id",
					Timestamp:          now.Add(-5 * time.Minute),
					OperatingState:     thermostat.CoolingOperatingState,
					CurrentTemperature: 18.8,
					CurrentHumidity:    &updatedCurrentHumidity,
				},
			},
		},
		{
			name: "should update current state without humidity",
			args: args{
				manager: &fakeCurrentStateManager{
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
						"operatingState": "COOLING",
						"currentTemperature": 18.8
					}`, now.Add(-5*time.Minute).Format(time.RFC3339Nano)),
				),
			},
			wantErr: false,
			wantManagerStates: map[string]thermostat.CurrentState{
				"test_device_id": {
					DeviceID:           "test_device_id",
					Timestamp:          now.Add(-5 * time.Minute),
					OperatingState:     thermostat.CoolingOperatingState,
					CurrentTemperature: 18.8,
				},
			},
		},
		{
			name: "should error if payload is invalid JSON",
			args: args{
				manager: &fakeCurrentStateManager{
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
			wantManagerStates: map[string]thermostat.CurrentState{
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
				manager: &fakeCurrentStateManager{
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
						"operatingState": "COLING",
						"currentTemperature": -100.0,
						"currentHumidity": 105.0
					}`, now.Add(-90*time.Minute).Format(time.RFC3339Nano)),
				),
			},
			wantErr: true,
			wantManagerStates: map[string]thermostat.CurrentState{
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
			name: "should error if already has newer state",
			args: args{
				manager: &fakeCurrentStateManager{
					States: map[string]thermostat.CurrentState{
						"test_device_id": {
							DeviceID:           "test_device_id",
							Timestamp:          now.Add(-5 * time.Minute),
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
						"operatingState": "COOLING",
						"currentTemperature": 18.8
					}`, now.Add(-30*time.Minute).Format(time.RFC3339Nano)),
				),
			},
			wantErr: true,
			wantManagerStates: map[string]thermostat.CurrentState{
				"test_device_id": {
					DeviceID:           "test_device_id",
					Timestamp:          now.Add(-5 * time.Minute),
					OperatingState:     thermostat.IdleOperatingState,
					CurrentTemperature: 17.5,
					CurrentHumidity:    &initialCurrentHumidity,
				},
			},
		},
		{
			name: "should error if failed to update",
			args: args{
				manager: &fakeCurrentStateManager{
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
						"operatingState": "COOLING",
						"currentTemperature": 18.8,
						"currentHumidity": 45.6
					}`, now.Add(-5*time.Minute).Format(time.RFC3339Nano)),
				),
			},
			wantErr: true,
			wantManagerStates: map[string]thermostat.CurrentState{
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
			handler := CurrentState(tt.args.manager)
			err := handler(context.Background(), tt.args.payload)

			// Check expected error
			if (err != nil) != tt.wantErr {
				t.Errorf("CurrentState() error = %v, wantErr %v", err, tt.wantErr)
			}

			// Check manager states
			if len(tt.args.manager.States) != len(tt.wantManagerStates) {
				t.Errorf("CurrentState() len(manager.States) = %d, want %d", len(tt.args.manager.States), len(tt.wantManagerStates))
			}
			for deviceID, wantState := range tt.wantManagerStates {
				state, exists := tt.args.manager.States[deviceID]
				if !exists {
					t.Errorf("CurrentState() manager.States[%s] not found", deviceID)
					continue
				}

				if !state.Timestamp.Equal(wantState.Timestamp) {
					t.Errorf("CurrentState() manager.States[%s].Timestamp = %v, want %v", deviceID, state.Timestamp, wantState.Timestamp)
				}

				if state.OperatingState != wantState.OperatingState {
					t.Errorf("CurrentState() manager.States[%s].OperationState = %v, want %v", deviceID, state.OperatingState, wantState.OperatingState)
				}

				if state.CurrentTemperature != wantState.CurrentTemperature {
					t.Errorf("CurrentState() manager.States[%s].TargetTemperature = %v, want %v", deviceID, state.CurrentTemperature, wantState.CurrentTemperature)
				}

				if !ptrEqual(state.CurrentHumidity, wantState.CurrentHumidity) {
					t.Errorf("CurrentState() manager.States[%s].CurrentHumidity = %v, want %v", deviceID, state.CurrentHumidity, wantState.CurrentHumidity)
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
