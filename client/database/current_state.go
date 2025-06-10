package database

import (
	"fmt"

	"github.com/alexchebotarsky/thermofridge-api/metrics"
	"github.com/alexchebotarsky/thermofridge-api/model/thermofridge"
)

const (
	OperatingStateKey     = "operatingState"
	CurrentTemperatureKey = "currentTemperature"
)

func (d *Database) FetchCurrentState() (*thermofridge.CurrentState, error) {
	var s thermofridge.CurrentState
	var err error

	operatingStateValue, err := d.GetStr(OperatingStateKey)
	if err != nil {
		return nil, fmt.Errorf("error getting %s from database: %v", OperatingStateKey, err)
	}
	operatingStateEnum := thermofridge.OperatingState(operatingStateValue)
	s.OperatingState = &operatingStateEnum

	currentTemperature, err := d.GetFloat(CurrentTemperatureKey)
	if err != nil {
		return nil, fmt.Errorf("error getting %s from database: %v", CurrentTemperatureKey, err)
	}
	s.CurrentTemperature = &currentTemperature

	return &s, nil
}

func (d *Database) UpdateCurrentState(state *thermofridge.CurrentState) (*thermofridge.CurrentState, error) {
	if state.OperatingState != nil {
		operatingState := *state.OperatingState
		err := d.Set(OperatingStateKey, string(operatingState))
		if err != nil {
			return nil, fmt.Errorf("error setting %s in database: %v", OperatingStateKey, err)
		}
		metrics.SetThermofridgeOperatingState(operatingState)
	}

	if state.CurrentTemperature != nil {
		temperature := *state.CurrentTemperature
		err := d.Set(CurrentTemperatureKey, fmt.Sprintf("%.2f", temperature))
		if err != nil {
			return nil, fmt.Errorf("error setting %s in database: %v", CurrentTemperatureKey, err)
		}
		metrics.SetThermofridgeCurrentTemperature(temperature)
	}

	return d.FetchCurrentState()
}
