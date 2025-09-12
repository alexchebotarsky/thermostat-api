package storage

import (
	"context"
	"testing"
	"time"

	"github.com/alexchebotarsky/thermostat-api/client"
	"github.com/alexchebotarsky/thermostat-api/model/thermostat"
)

const testDeviceID = "test-device-id"
const defaultTargetTemperature = 20
const defaultMode = thermostat.OffMode

func newTestStorage(ctx context.Context, t *testing.T) *Client {
	dbPath := ":memory:" // SQLite in-memory
	s, err := New(ctx, dbPath, defaultMode, defaultTargetTemperature)
	if err != nil {
		t.Fatalf("Error creating new storage: %v", err)
	}
	t.Cleanup(func() { _ = s.Close() })
	return s
}

func compareTargetStates(t *testing.T, got, want *thermostat.TargetState) {
	if got.DeviceID != want.DeviceID {
		t.Errorf("DeviceID = %v, want %v", got.DeviceID, want.DeviceID)
	}

	if !ptrEqual(got.Mode, want.Mode) {
		t.Errorf("Mode = %v, want %v", got.Mode, want.Mode)
	}

	if !ptrEqual(got.TargetTemperature, want.TargetTemperature) {
		t.Errorf("TargetTemperature = %v, want %v", got.TargetTemperature, want.TargetTemperature)
	}
}

func compareCurrentStates(t *testing.T, got, want *thermostat.CurrentState) {
	if got.DeviceID != want.DeviceID {
		t.Errorf("DeviceID = %v, want %v", got.DeviceID, want.DeviceID)
	}

	if !got.Timestamp.Equal(want.Timestamp) {
		t.Errorf("Timestamp = %v, want %v", got.Timestamp, want.Timestamp)
	}

	if got.OperatingState != want.OperatingState {
		t.Errorf("OperatingState = %v, want %v", got.OperatingState, want.OperatingState)
	}

	if got.CurrentTemperature != want.CurrentTemperature {
		t.Errorf("CurrentTemperature = %v, want %v", got.CurrentTemperature, want.CurrentTemperature)
	}

	if !ptrEqual(got.CurrentHumidity, want.CurrentHumidity) {
		t.Errorf("CurrentHumidity = %v, want %v", got.CurrentHumidity, want.CurrentHumidity)
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

func TestTargetStateIntegration(t *testing.T) {
	ctx := context.Background()
	s := newTestStorage(ctx, t)

	defaultMode := defaultMode
	defaultTargetTemperature := defaultTargetTemperature
	defaultState := &thermostat.TargetState{
		DeviceID:          testDeviceID,
		Mode:              &defaultMode,
		TargetTemperature: &defaultTargetTemperature,
	}

	initialMode := thermostat.AutoMode
	initialTargetTemperature := 22
	initialState := &thermostat.TargetState{
		DeviceID:          testDeviceID,
		Mode:              &initialMode,
		TargetTemperature: &initialTargetTemperature,
	}

	updatedMode := thermostat.HeatMode
	updatedTargetTemperature := 20
	updatedState := &thermostat.TargetState{
		DeviceID:          testDeviceID,
		Mode:              &updatedMode,
		TargetTemperature: &updatedTargetTemperature,
	}

	// Read (defaults)
	got, err := s.FetchTargetState(ctx, testDeviceID)
	if err != nil {
		t.Fatalf("Error fetching default target state: %v", err)
	}

	compareTargetStates(t, got, defaultState)

	// Create
	got, err = s.UpdateTargetState(ctx, initialState)
	if err != nil {
		t.Fatalf("Error creating target state: %v", err)
	}

	compareTargetStates(t, got, initialState)

	// Read (created)
	got, err = s.FetchTargetState(ctx, testDeviceID)
	if err != nil {
		t.Fatalf("Error reading target state: %v", err)
	}

	compareTargetStates(t, got, initialState)

	// Update
	got, err = s.UpdateTargetState(ctx, updatedState)
	if err != nil {
		t.Fatalf("Error updating target state: %v", err)
	}

	compareTargetStates(t, got, updatedState)

	// Read (updated)
	got, err = s.FetchTargetState(ctx, testDeviceID)
	if err != nil {
		t.Fatalf("Error reading target state: %v", err)
	}

	compareTargetStates(t, got, updatedState)
}

func TestCurrentStateIntegration(t *testing.T) {
	ctx := context.Background()
	s := newTestStorage(ctx, t)

	currentHumidity := 45.2
	state := &thermostat.CurrentState{
		DeviceID:           testDeviceID,
		Timestamp:          time.Now().Add(-5 * time.Minute),
		OperatingState:     thermostat.CoolingOperatingState,
		CurrentTemperature: 22.5,
		CurrentHumidity:    &currentHumidity,
	}

	updatedHumidity := 47.8
	updatedState := &thermostat.CurrentState{
		DeviceID:           testDeviceID,
		Timestamp:          time.Now(),
		OperatingState:     thermostat.HeatingOperatingState,
		CurrentTemperature: 19.0,
		CurrentHumidity:    &updatedHumidity,
	}

	// Read (not found)
	_, err := s.FetchCurrentState(ctx, testDeviceID)
	switch err.(type) {
	case *client.ErrNotFound:
		// Expected
	case nil:
		t.Fatal("Expected error when fetching non-existent current state, got nil")
	default:
		t.Fatalf("Expected ErrNotFound when fetching non-existent current state, got: %v", err)
	}

	// Create
	got, err := s.UpdateCurrentState(ctx, state)
	if err != nil {
		t.Fatalf("Error creating current state: %v", err)
	}

	compareCurrentStates(t, got, state)

	// Read (created)
	got, err = s.FetchCurrentState(ctx, testDeviceID)
	if err != nil {
		t.Fatalf("Error reading current state: %v", err)
	}

	compareCurrentStates(t, got, state)

	// Update
	got, err = s.UpdateCurrentState(ctx, updatedState)
	if err != nil {
		t.Fatalf("Error updating current state: %v", err)
	}

	compareCurrentStates(t, got, updatedState)

	// Read (updated)
	got, err = s.FetchCurrentState(ctx, testDeviceID)
	if err != nil {
		t.Fatalf("Error reading updated current state: %v", err)
	}

	compareCurrentStates(t, got, updatedState)
}
