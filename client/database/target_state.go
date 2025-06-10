package database

import (
	"fmt"

	"github.com/alexchebotarsky/thermofridge-api/metrics"
	"github.com/alexchebotarsky/thermofridge-api/model/thermofridge"
)

const (
	ModeKey              = "mode"
	TargetTemperatureKey = "targetTemperature"
)

func (d *Database) prepareTargetState() error {
	mode, err := d.GetStr(ModeKey)
	if err != nil {
		return fmt.Errorf("error getting initial %s from database: %v", ModeKey, err)
	}
	metrics.SetThermofridgeMode(thermofridge.Mode(mode))

	targetTemperature, err := d.GetInt(TargetTemperatureKey)
	if err != nil {
		return fmt.Errorf("error getting initial %s from database: %v", TargetTemperatureKey, err)
	}
	metrics.SetThermofridgeTargetTemperature(targetTemperature)

	return nil
}

func (d *Database) FetchTargetState() (*thermofridge.TargetState, error) {
	var s thermofridge.TargetState
	var err error

	modeValue, err := d.GetStr(ModeKey)
	if err != nil {
		return nil, fmt.Errorf("error getting %s from database: %v", ModeKey, err)
	}
	modeEnum := thermofridge.Mode(modeValue)
	s.Mode = &modeEnum

	targetTemperature, err := d.GetInt(TargetTemperatureKey)
	if err != nil {
		return nil, fmt.Errorf("error getting %s from database: %v", TargetTemperatureKey, err)
	}
	s.TargetTemperature = &targetTemperature

	return &s, nil
}

func (d *Database) UpdateTargetState(state *thermofridge.TargetState) (*thermofridge.TargetState, error) {
	if state.Mode != nil {
		mode := *state.Mode
		err := d.Set(ModeKey, string(mode))
		if err != nil {
			return nil, fmt.Errorf("error setting %s in database: %v", ModeKey, err)
		}
		metrics.SetThermofridgeMode(mode)
	}

	if state.TargetTemperature != nil {
		temperature := *state.TargetTemperature
		err := d.Set(TargetTemperatureKey, fmt.Sprintf("%d", temperature))
		if err != nil {
			return nil, fmt.Errorf("error setting %s in database: %v", TargetTemperatureKey, err)
		}
		metrics.SetThermofridgeTargetTemperature(temperature)
	}

	return d.FetchTargetState()
}
