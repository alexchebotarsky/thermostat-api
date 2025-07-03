package thermofridge

import "fmt"

type TargetState struct {
	DeviceID          string `json:"deviceID" db:"device_id"`
	Mode              *Mode  `json:"mode" db:"mode"`
	TargetTemperature *int   `json:"targetTemperature" db:"target_temperature"`
}

func (s *TargetState) Validate() error {
	if s.Mode != nil {
		switch *s.Mode {
		case OffMode, HeatMode, CoolMode, AutoMode:
			// Valid
		default:
			return fmt.Errorf("mode must be one of: [%s, %s, %s, %s], got: %q", OffMode, HeatMode, CoolMode, AutoMode, *s.Mode)
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
