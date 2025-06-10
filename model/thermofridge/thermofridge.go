package thermofridge

import (
	"fmt"
)

type TargetState struct {
	Mode              *Mode `json:"mode"`
	TargetTemperature *int  `json:"targetTemperature"`
}

func (s *TargetState) Validate() error {
	if s.Mode != nil {
		switch *s.Mode {
		case OffMode, HeatMode, CoolMode, AutoMode:
			// Valid
		default:
			return fmt.Errorf("mode must be one of: [%s, %s, %s, %s], got: %s", OffMode, HeatMode, CoolMode, AutoMode, *s.Mode)
		}
	}

	if s.TargetTemperature != nil {
		if *s.TargetTemperature < 0 || *s.TargetTemperature > 25 {
			return fmt.Errorf("target temperature must be in range [0,25]. got: %d", *s.TargetTemperature)
		}
	}

	return nil
}

type Mode string

const (
	OffMode  Mode = "OFF"
	HeatMode Mode = "HEAT"
	CoolMode Mode = "COOL"
	AutoMode Mode = "AUTO"
)

type CurrentState struct {
	OperatingState     *OperatingState `json:"operatingState"`
	CurrentTemperature *float64        `json:"currentTemperature"`
}

func (s *CurrentState) Validate() error {
	if s.OperatingState != nil {
		switch *s.OperatingState {
		case IdleOperatingState, HeatingOperatingState, CoolingOperatingState:
			// Valid
		default:
			return fmt.Errorf("operating state must be one of: [%s, %s, %s], got: %s", IdleOperatingState, HeatingOperatingState, CoolingOperatingState, *s.OperatingState)
		}
	}

	if s.CurrentTemperature != nil {
		if *s.CurrentTemperature < -55 || *s.CurrentTemperature > 125 {
			return fmt.Errorf("current temperature must be in range [-55,125]. got: %f", *s.CurrentTemperature)
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

type TemperatureReading struct {
	Temperature float64 `json:"temperature"`
}
