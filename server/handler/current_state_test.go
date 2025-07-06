package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/alexchebotarsky/thermostat-api/client"
	"github.com/alexchebotarsky/thermostat-api/model/thermostat"
)

type fakeCurrentStateFetcher struct {
	States map[string]thermostat.CurrentState

	shouldFail bool
}

func (f *fakeCurrentStateFetcher) FetchCurrentState(ctx context.Context, deviceID string) (*thermostat.CurrentState, error) {
	if f.shouldFail {
		return nil, errors.New("test error")
	}

	state, exists := f.States[deviceID]
	if !exists {
		return nil, &client.ErrNotFound{Err: errors.New("state for the device not found")}
	}

	return &state, nil
}

func TestGetCurrentState(t *testing.T) {
	now := time.Now()

	type args struct {
		fetcher *fakeCurrentStateFetcher
		req     *http.Request
	}
	tests := []struct {
		name string
		args args
		// HTTP response expectations
		wantStatus int
		wantErr    bool
		wantBody   *thermostat.CurrentState
	}{
		{
			name: "should fetch current state",
			args: args{
				fetcher: &fakeCurrentStateFetcher{
					States: map[string]thermostat.CurrentState{
						"test_device_id": {
							DeviceID:           "test_device_id",
							OperatingState:     thermostat.HeatingOperatingState,
							CurrentTemperature: 15.0,
							Timestamp:          now.Add(-5 * time.Minute),
						},
					},
					shouldFail: false,
				},
				req: addChiURLParams(
					httptest.NewRequest(http.MethodGet, "/api/v1/current-state/test_device_id", nil),
					map[string]string{"deviceID": "test_device_id"},
				),
			},
			wantStatus: http.StatusOK,
			wantErr:    false,
			wantBody: &thermostat.CurrentState{
				DeviceID:           "test_device_id",
				OperatingState:     thermostat.HeatingOperatingState,
				CurrentTemperature: 15.0,
				Timestamp:          now.Add(-5 * time.Minute),
			},
		},
		{
			name: "should return error 404, if current state not found",
			args: args{
				fetcher: &fakeCurrentStateFetcher{
					States:     map[string]thermostat.CurrentState{},
					shouldFail: false,
				},
				req: addChiURLParams(
					httptest.NewRequest(http.MethodGet, "/api/v1/current-state/test_device_id", nil),
					map[string]string{"deviceID": "test_device_id"},
				),
			},
			wantStatus: http.StatusNotFound,
			wantErr:    true,
			wantBody:   nil,
		},
		{
			name: "should return error 500, if failed to fetch",
			args: args{
				fetcher: &fakeCurrentStateFetcher{
					States: map[string]thermostat.CurrentState{
						"test_device_id": {
							DeviceID:           "test_device_id",
							OperatingState:     thermostat.HeatingOperatingState,
							CurrentTemperature: 15.0,
							Timestamp:          now.Add(-5 * time.Minute),
						},
					},
					shouldFail: true,
				},
				req: addChiURLParams(
					httptest.NewRequest(http.MethodGet, "/api/v1/current-state/test_device_id", nil),
					map[string]string{"deviceID": "test_device_id"},
				),
			},
			wantStatus: http.StatusInternalServerError,
			wantErr:    true,
			wantBody:   nil,
		},
		{
			name: "should return error 500, if current state is invalid",
			args: args{
				fetcher: &fakeCurrentStateFetcher{
					States: map[string]thermostat.CurrentState{
						"test_device_id": {
							DeviceID:           "test_device_id",
							OperatingState:     thermostat.HeatingOperatingState,
							CurrentTemperature: -100.0,
							Timestamp:          now.Add(-90 * time.Minute),
						},
					},
					shouldFail: false,
				},
				req: addChiURLParams(
					httptest.NewRequest(http.MethodGet, "/api/v1/current-state/test_device_id", nil),
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
			handler := GetCurrentState(tt.args.fetcher)
			handler(w, tt.args.req)

			// Check the status code
			if w.Code != tt.wantStatus {
				t.Errorf("GetCurrentState() status = %v, want %v", w.Code, tt.wantStatus)
			}

			// If we expect an error, we just check that response body is not empty and return early
			if tt.wantErr {
				if w.Body.Len() == 0 {
					t.Errorf("GetCurrentState() response body is empty, want error")
				}
				return
			}

			// Decode the response body into struct for checking
			var resBody thermostat.CurrentState
			if err := json.NewDecoder(w.Body).Decode(&resBody); err != nil {
				t.Fatalf("GetCurrentState() error json decoding response body: %v", err)
			}

			// Check response body fields
			if resBody.DeviceID != tt.wantBody.DeviceID {
				t.Errorf("GetCurrentState() response deviceID = %v, want %v", resBody.DeviceID, tt.wantBody.DeviceID)
			}
			if resBody.OperatingState != tt.wantBody.OperatingState {
				t.Errorf("GetCurrentState() response operatingState = %v, want %v", resBody.OperatingState, tt.wantBody.OperatingState)
			}
			if resBody.CurrentTemperature != tt.wantBody.CurrentTemperature {
				t.Errorf("GetCurrentState() response currentTemperature = %v, want %v", resBody.CurrentTemperature, tt.wantBody.CurrentTemperature)
			}
			if !resBody.Timestamp.Equal(tt.wantBody.Timestamp) {
				t.Errorf("GetCurrentState() response timestamp = %v, want %v", resBody.Timestamp, tt.wantBody.Timestamp)
			}
		})
	}
}
