package thermostat

import (
	"fmt"
	"time"
)

type CurrentState struct {
	DeviceID           string         `json:"deviceID" db:"device_id"`
	Timestamp          time.Time      `json:"timestamp" db:"timestamp"`
	OperatingState     OperatingState `json:"operatingState" db:"operating_state"`
	CurrentTemperature float64        `json:"currentTemperature" db:"current_temperature"`
	CurrentHumidity    *float64       `json:"currentHumidity,omitempty" db:"current_humidity"` // Not all thermostats may report humidity
}

func (s *CurrentState) Validate() error {
	if s.DeviceID == "" {
		return fmt.Errorf("device ID cannot be empty")
	}

	if time.Since(s.Timestamp) > 1*time.Hour {
		return fmt.Errorf("timestamp cannot be older than 1 hour, got: '%s'", s.Timestamp)
	}

	switch s.OperatingState {
	case IdleOperatingState, HeatingOperatingState, CoolingOperatingState:
		// Valid
	default:
		return fmt.Errorf("operating state must be one of: [%s, %s, %s], got: '%s'", IdleOperatingState, HeatingOperatingState, CoolingOperatingState, s.OperatingState)
	}

	if s.CurrentTemperature < -55 || s.CurrentTemperature > 125 {
		return fmt.Errorf("current temperature must be in range [-55,125]. got: %.2f", s.CurrentTemperature)
	}

	if s.CurrentHumidity != nil {
		if *s.CurrentHumidity < 0 || *s.CurrentHumidity > 100 {
			return fmt.Errorf("current humidity must be in range [0,100]. got: %.2f", *s.CurrentHumidity)
		}
	}

	return nil
}

type OperatingState string

const (
	IdleOperatingState    OperatingState = "IDLE"
	HeatingOperatingState OperatingState = "HEATING"
	CoolingOperatingState OperatingState = "COOLING"
)
