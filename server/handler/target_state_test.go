package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alexchebotarsky/thermofridge-api/client"
	"github.com/alexchebotarsky/thermofridge-api/model/thermofridge"
)

type fakeTargetStateFetcher struct {
	States map[string]thermofridge.TargetState

	shouldFail bool
}

func (f *fakeTargetStateFetcher) FetchTargetState(ctx context.Context, deviceID string) (*thermofridge.TargetState, error) {
	if f.shouldFail {
		return nil, errors.New("test error")
	}

	state, exists := f.States[deviceID]
	if !exists {
		return nil, &client.ErrNotFound{Err: errors.New("state for the device not found")}
	}

	return &state, nil
}

func TestGetTargetState(t *testing.T) {
	testMode := thermofridge.HeatMode
	testTargetTemperature := 25

	type args struct {
		fetcher *fakeTargetStateFetcher
		req     *http.Request
	}
	tests := []struct {
		name string
		args args
		// HTTP response expectations
		wantStatus int
		wantErr    bool
		wantBody   *thermofridge.TargetState
	}{
		{
			name: "should fetch target state",
			args: args{
				fetcher: &fakeTargetStateFetcher{
					States: map[string]thermofridge.TargetState{
						"test_device_id": {
							DeviceID:          "test_device_id",
							Mode:              &testMode,
							TargetTemperature: &testTargetTemperature,
						},
					},
					shouldFail: false,
				},
				req: addChiURLParams(
					httptest.NewRequest(http.MethodGet, "/api/v1/target-state/test_device_id", nil),
					map[string]string{"deviceID": "test_device_id"},
				),
			},
			wantStatus: http.StatusOK,
			wantErr:    false,
			wantBody: &thermofridge.TargetState{
				DeviceID:          "test_device_id",
				Mode:              &testMode,
				TargetTemperature: &testTargetTemperature,
			},
		},
		{
			name: "should return error 500, if failed to fetch",
			args: args{
				fetcher: &fakeTargetStateFetcher{
					States: map[string]thermofridge.TargetState{
						"test_device_id": {
							DeviceID:          "test_device_id",
							Mode:              &testMode,
							TargetTemperature: &testTargetTemperature,
						},
					},
					shouldFail: true,
				},
				req: addChiURLParams(
					httptest.NewRequest(http.MethodGet, "/api/v1/target-state/test_device_id", nil),
					map[string]string{"deviceID": "test_device_id"},
				),
			},
			wantStatus: http.StatusInternalServerError,
			wantErr:    true,
			wantBody:   nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			handler := GetTargetState(tt.args.fetcher)
			handler(w, tt.args.req)

			// Check the status code
			if w.Code != tt.wantStatus {
				t.Errorf("GetTargetState() status = %v, want %v", w.Code, tt.wantStatus)
			}

			// If we expect an error, we just check that response body is not empty and return early
			if tt.wantErr {
				if w.Body.Len() == 0 {
					t.Errorf("GetTargetState() response body is empty, want error")
				}
				return
			}

			// Decode the response body into struct for checking
			var resBody thermofridge.TargetState
			if err := json.NewDecoder(w.Body).Decode(&resBody); err != nil {
				t.Fatalf("GetTargetState() error json decoding response body: %v", err)
			}

			// Check response body fields
			if resBody.DeviceID != tt.wantBody.DeviceID {
				t.Errorf("GetTargetState() response body DeviceID = %v, want %v", resBody.DeviceID, tt.wantBody.DeviceID)
			}
			if !ptrEqual(resBody.Mode, tt.wantBody.Mode) {
				t.Errorf("GetTargetState() response body Mode = %v, want %v", resBody.Mode, tt.wantBody.Mode)
			}
			if !ptrEqual(resBody.TargetTemperature, tt.wantBody.TargetTemperature) {
				t.Errorf("GetTargetState() response body TargetTemperature = %v, want %v", resBody.TargetTemperature, tt.wantBody.TargetTemperature)
			}
		})
	}
}

type fakeTargetStateUpdater struct {
	States map[string]thermofridge.TargetState

	shouldFail bool
}

func (f *fakeTargetStateUpdater) UpdateTargetState(ctx context.Context, state *thermofridge.TargetState) (*thermofridge.TargetState, error) {
	if f.shouldFail {
		return nil, errors.New("test error")
	}

	if state != nil {
		oldState, exists := f.States[state.DeviceID]
		if !exists {
			f.States[state.DeviceID] = *state
		} else {
			if state.Mode != nil {
				oldState.Mode = state.Mode
			}
			if state.TargetTemperature != nil {
				oldState.TargetTemperature = state.TargetTemperature
			}
			f.States[state.DeviceID] = oldState
		}
	}

	updatedState := f.States[state.DeviceID]

	return &updatedState, nil
}

type fakeTargetStatePublisher struct {
	States []thermofridge.TargetState

	shouldFail bool
}

func (m *fakeTargetStatePublisher) PublishTargetState(ctx context.Context, state *thermofridge.TargetState) error {
	if m.shouldFail {
		return errors.New("test error")
	}

	if state != nil {
		m.States = append(m.States, *state)
	}

	return nil
}

func TestUpdateTargetState(t *testing.T) {
	initialMode := thermofridge.HeatMode
	initialTargetTemperature := 25
	updatedMode := thermofridge.CoolMode
	updatedTargetTemperature := 15
	invalidMode := thermofridge.Mode("INVALID_MODE")
	invalidTargetTemperature := -5

	type args struct {
		updater   *fakeTargetStateUpdater
		publisher *fakeTargetStatePublisher
		req       *http.Request
	}
	tests := []struct {
		name string
		args args
		// HTTP response expectations
		wantStatus int
		wantErr    bool
		wantBody   *thermofridge.TargetState
		// Updater expectations
		wantUpdaterStates map[string]*thermofridge.TargetState
		// Publisher expectations
		wantPublisherStates []*thermofridge.TargetState
	}{
		{
			name: "should update target state and publish updated state",
			args: args{
				updater: &fakeTargetStateUpdater{
					States: map[string]thermofridge.TargetState{
						"test_device_id": {
							DeviceID:          "test_device_id",
							Mode:              &initialMode,
							TargetTemperature: &initialTargetTemperature,
						},
					},
					shouldFail: false,
				},
				publisher: &fakeTargetStatePublisher{
					States:     []thermofridge.TargetState{},
					shouldFail: false,
				},
				req: addChiURLParams(
					httptest.NewRequest(http.MethodPost, "/api/v1/target-state/test_device_id", bytes.NewReader(
						[]byte(fmt.Sprintf(`{
							"mode": "%s",
							"targetTemperature": %d
						}`, updatedMode, updatedTargetTemperature))),
					),
					map[string]string{"deviceID": "test_device_id"},
				),
			},
			wantStatus: http.StatusOK,
			wantErr:    false,
			wantBody: &thermofridge.TargetState{
				DeviceID:          "test_device_id",
				Mode:              &updatedMode,
				TargetTemperature: &updatedTargetTemperature,
			},
			wantUpdaterStates: map[string]*thermofridge.TargetState{
				"test_device_id": {
					DeviceID:          "test_device_id",
					Mode:              &updatedMode,
					TargetTemperature: &updatedTargetTemperature,
				},
			},
			wantPublisherStates: []*thermofridge.TargetState{
				{
					DeviceID:          "test_device_id",
					Mode:              &updatedMode,
					TargetTemperature: &updatedTargetTemperature,
				},
			},
		},
		{
			name: "should update only mode and publish updated state",
			args: args{
				updater: &fakeTargetStateUpdater{
					States: map[string]thermofridge.TargetState{
						"test_device_id": {
							DeviceID:          "test_device_id",
							Mode:              &initialMode,
							TargetTemperature: &initialTargetTemperature,
						},
					},
					shouldFail: false,
				},
				publisher: &fakeTargetStatePublisher{
					States:     []thermofridge.TargetState{},
					shouldFail: false,
				},
				req: addChiURLParams(
					httptest.NewRequest(http.MethodPost, "/api/v1/target-state/test_device_id", bytes.NewReader(
						[]byte(fmt.Sprintf(`{
							"mode": "%s"
						}`, updatedMode))),
					),
					map[string]string{"deviceID": "test_device_id"},
				),
			},
			wantStatus: http.StatusOK,
			wantErr:    false,
			wantBody: &thermofridge.TargetState{
				DeviceID:          "test_device_id",
				Mode:              &updatedMode,
				TargetTemperature: &initialTargetTemperature,
			},
			wantUpdaterStates: map[string]*thermofridge.TargetState{
				"test_device_id": {
					DeviceID:          "test_device_id",
					Mode:              &updatedMode,
					TargetTemperature: &initialTargetTemperature,
				},
			},
			wantPublisherStates: []*thermofridge.TargetState{
				{
					DeviceID:          "test_device_id",
					Mode:              &updatedMode,
					TargetTemperature: &initialTargetTemperature,
				},
			},
		},
		{
			name: "should update only target temperature and publish updated state",
			args: args{
				updater: &fakeTargetStateUpdater{
					States: map[string]thermofridge.TargetState{
						"test_device_id": {
							DeviceID:          "test_device_id",
							Mode:              &initialMode,
							TargetTemperature: &initialTargetTemperature,
						},
					},
					shouldFail: false,
				},
				publisher: &fakeTargetStatePublisher{
					States:     []thermofridge.TargetState{},
					shouldFail: false,
				},
				req: addChiURLParams(
					httptest.NewRequest(http.MethodPost, "/api/v1/target-state/test_device_id", bytes.NewReader(
						[]byte(fmt.Sprintf(`{
							"targetTemperature": %d
						}`, updatedTargetTemperature))),
					),
					map[string]string{"deviceID": "test_device_id"},
				),
			},
			wantStatus: http.StatusOK,
			wantErr:    false,
			wantBody: &thermofridge.TargetState{
				DeviceID:          "test_device_id",
				Mode:              &initialMode,
				TargetTemperature: &updatedTargetTemperature,
			},
			wantUpdaterStates: map[string]*thermofridge.TargetState{
				"test_device_id": {
					DeviceID:          "test_device_id",
					Mode:              &initialMode,
					TargetTemperature: &updatedTargetTemperature,
				},
			},
			wantPublisherStates: []*thermofridge.TargetState{
				{
					DeviceID:          "test_device_id",
					Mode:              &initialMode,
					TargetTemperature: &updatedTargetTemperature,
				},
			},
		},
		{
			name: "should return error 400, if request body is invalid JSON",
			args: args{
				updater: &fakeTargetStateUpdater{
					States: map[string]thermofridge.TargetState{
						"test_device_id": {
							DeviceID:          "test_device_id",
							Mode:              &initialMode,
							TargetTemperature: &initialTargetTemperature,
						},
					},
					shouldFail: false,
				},
				publisher: &fakeTargetStatePublisher{
					States:     []thermofridge.TargetState{},
					shouldFail: false,
				},
				req: addChiURLParams(
					httptest.NewRequest(http.MethodPost, "/api/v1/target-state/test_device_id", bytes.NewReader(
						[]byte(`{
							"mode":
						}`)),
					),
					map[string]string{"deviceID": "test_device_id"},
				),
			},
			wantStatus: http.StatusBadRequest,
			wantErr:    true,
		},
		{
			name: "should return error 400, if request body has invalid values",
			args: args{
				updater: &fakeTargetStateUpdater{
					States: map[string]thermofridge.TargetState{
						"test_device_id": {
							DeviceID:          "test_device_id",
							Mode:              &initialMode,
							TargetTemperature: &initialTargetTemperature,
						},
					},
					shouldFail: false,
				},
				publisher: &fakeTargetStatePublisher{
					States:     []thermofridge.TargetState{},
					shouldFail: false,
				},
				req: addChiURLParams(
					httptest.NewRequest(http.MethodPost, "/api/v1/target-state/test_device_id", bytes.NewReader(
						[]byte(fmt.Sprintf(`{
							"mode": "%s",
							"targetTemperature": %d
						}`, invalidMode, invalidTargetTemperature))),
					),
					map[string]string{"deviceID": "test_device_id"},
				),
			},
			wantStatus: http.StatusBadRequest,
			wantErr:    true,
		},
		{
			name: "should return error 500, if failed to update",
			args: args{
				updater: &fakeTargetStateUpdater{
					States: map[string]thermofridge.TargetState{
						"test_device_id": {
							DeviceID:          "test_device_id",
							Mode:              &initialMode,
							TargetTemperature: &initialTargetTemperature,
						},
					},
					shouldFail: true,
				},
				publisher: &fakeTargetStatePublisher{
					States:     []thermofridge.TargetState{},
					shouldFail: false,
				},
				req: addChiURLParams(
					httptest.NewRequest(http.MethodPost, "/api/v1/target-state/test_device_id", bytes.NewReader(
						[]byte(fmt.Sprintf(`{
							"mode": "%s",
							"targetTemperature": %d
						}`, updatedMode, updatedTargetTemperature))),
					),
					map[string]string{"deviceID": "test_device_id"},
				),
			},
			wantStatus: http.StatusInternalServerError,
			wantErr:    true,
		},
		{
			name: "should return error 500, if failed to publish",
			args: args{
				updater: &fakeTargetStateUpdater{
					States: map[string]thermofridge.TargetState{
						"test_device_id": {
							DeviceID:          "test_device_id",
							Mode:              &initialMode,
							TargetTemperature: &initialTargetTemperature,
						},
					},
					shouldFail: false,
				},
				publisher: &fakeTargetStatePublisher{
					States:     []thermofridge.TargetState{},
					shouldFail: true,
				},
				req: addChiURLParams(
					httptest.NewRequest(http.MethodPost, "/api/v1/target-state/test_device_id", bytes.NewReader(
						[]byte(fmt.Sprintf(`{
							"mode": "%s",
							"targetTemperature": %d
						}`, updatedMode, updatedTargetTemperature))),
					),
					map[string]string{"deviceID": "test_device_id"},
				),
			},
			wantStatus: http.StatusInternalServerError,
			wantErr:    true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			handler := UpdateTargetState(tt.args.updater, tt.args.publisher)
			handler(w, tt.args.req)

			// Check response status code
			if w.Code != tt.wantStatus {
				t.Errorf("UpdateTargetState() status = %v, want %v", w.Code, tt.wantStatus)
			}

			// If we expect an error, we just check that response body is not empty and return early
			if tt.wantErr {
				if w.Body.Len() == 0 {
					t.Errorf("UpdateTargetState() response body is empty, want error")
				}
				return
			}

			// Decode the response body into struct for checking
			var resBody thermofridge.TargetState
			if err := json.NewDecoder(w.Body).Decode(&resBody); err != nil {
				t.Fatalf("UpdateTargetState() error json decoding response body: %v", err)
			}

			// Check response body fields
			if resBody.DeviceID != tt.wantBody.DeviceID {
				t.Errorf("UpdateTargetState() response body DeviceID = %v, want %v", resBody.DeviceID, tt.wantBody.DeviceID)
			}
			if !ptrEqual(resBody.Mode, tt.wantBody.Mode) {
				t.Errorf("UpdateTargetState() response body Mode = %v, want %v", resBody.Mode, tt.wantBody.Mode)
			}
			if !ptrEqual(resBody.TargetTemperature, tt.wantBody.TargetTemperature) {
				t.Errorf("UpdateTargetState() response body TargetTemperature = %v, want %v", resBody.TargetTemperature, tt.wantBody.TargetTemperature)
			}

			// Check updater states
			if len(tt.args.updater.States) != len(tt.wantUpdaterStates) {
				t.Errorf("UpdateTargetState() len(updater.States) = %d, want %d", len(tt.args.updater.States), len(tt.wantUpdaterStates))
			}
			for deviceID, wantState := range tt.wantUpdaterStates {
				state, exists := tt.args.updater.States[deviceID]
				if !exists {
					t.Errorf("UpdateTargetState() updater.States[%q] not found", deviceID)
					continue
				}

				if !ptrEqual(state.Mode, wantState.Mode) {
					t.Errorf("UpdateTargetState() updater.States[%q].Mode = %v, want %v", deviceID, state.Mode, wantState.Mode)
				}

				if !ptrEqual(state.TargetTemperature, wantState.TargetTemperature) {
					t.Errorf("UpdateTargetState() updater.States[%q].TargetTemperature = %v, want %v", deviceID, state.TargetTemperature, wantState.TargetTemperature)
				}
			}

			// Check publisher states
			if len(tt.args.publisher.States) != len(tt.wantPublisherStates) {
				t.Errorf("UpdateTargetState() len(publisher.States) = %d, want %d", len(tt.args.publisher.States), len(tt.wantPublisherStates))
			}
			for i, state := range tt.args.publisher.States {
				wantState := tt.wantPublisherStates[i]

				if !ptrEqual(state.Mode, wantState.Mode) {
					t.Errorf("UpdateTargetState() publisher.States[%d].Mode = %v, want %v", i, state.Mode, wantState.Mode)
				}

				if !ptrEqual(state.TargetTemperature, wantState.TargetTemperature) {
					t.Errorf("UpdateTargetState() publisher.States[%d].TargetTemperature = %v, want %v", i, state.TargetTemperature, wantState.TargetTemperature)
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
