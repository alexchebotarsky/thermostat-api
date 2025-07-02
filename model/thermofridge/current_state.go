package thermofridge

import "fmt"

type CurrentState struct {
	DeviceID           string         `json:"deviceID" db:"device_id"`
	OperatingState     OperatingState `json:"operatingState" db:"operating_state"`
	CurrentTemperature float64        `json:"currentTemperature" db:"current_temperature"`
}

func (s *CurrentState) Validate() error {
	switch s.OperatingState {
	case IdleOperatingState, HeatingOperatingState, CoolingOperatingState:
		// Valid
	default:
		return fmt.Errorf("operating state must be one of: [%s, %s, %s], got: %s", IdleOperatingState, HeatingOperatingState, CoolingOperatingState, s.OperatingState)
	}

	if s.CurrentTemperature < -55 || s.CurrentTemperature > 125 {
		return fmt.Errorf("current temperature must be in range [-55,125]. got: %f", s.CurrentTemperature)
	}

	return nil
}

type OperatingState string

const (
	IdleOperatingState    OperatingState = "IDLE"
	HeatingOperatingState OperatingState = "HEATING"
	CoolingOperatingState OperatingState = "COOLING"
)
