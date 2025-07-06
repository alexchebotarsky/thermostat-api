package thermostat

import (
	"fmt"
	"time"
)

type CurrentState struct {
	DeviceID           string         `json:"deviceID" db:"device_id"`
	OperatingState     OperatingState `json:"operatingState" db:"operating_state"`
	CurrentTemperature float64        `json:"currentTemperature" db:"current_temperature"`
	Timestamp          time.Time      `json:"timestamp" db:"timestamp"`
}

func (s *CurrentState) Validate() error {
	if s.DeviceID == "" {
		return fmt.Errorf("device ID cannot be empty")
	}

	switch s.OperatingState {
	case IdleOperatingState, HeatingOperatingState, CoolingOperatingState:
		// Valid
	default:
		return fmt.Errorf("operating state must be one of: [%s, %s, %s], got: %q", IdleOperatingState, HeatingOperatingState, CoolingOperatingState, s.OperatingState)
	}

	if s.CurrentTemperature < -55 || s.CurrentTemperature > 125 {
		return fmt.Errorf("current temperature must be in range [-55,125]. got: %.2f", s.CurrentTemperature)
	}

	if time.Since(s.Timestamp) > 1*time.Hour {
		return fmt.Errorf("timestamp cannot be older than 1 hour, got: %q", s.Timestamp)
	}

	return nil
}

type OperatingState string

const (
	IdleOperatingState    OperatingState = "IDLE"
	HeatingOperatingState OperatingState = "HEATING"
	CoolingOperatingState OperatingState = "COOLING"
)
