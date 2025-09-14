package storage

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/alexchebotarsky/thermostat-api/metrics"
	"github.com/alexchebotarsky/thermostat-api/model/thermostat"
)

func (c *Client) initTargetStateTable(ctx context.Context) error {
	schema := `
		CREATE TABLE IF NOT EXISTS target_state (
			device_id TEXT PRIMARY KEY,
			mode TEXT,
			target_temperature INTEGER
		);
	`

	_, err := c.db.ExecContext(ctx, schema)
	if err != nil {
		return fmt.Errorf("error executing target state schema: %v", err)
	}

	return nil
}

func (c *Client) reportTargetStateMetrics(ctx context.Context) error {
	query := `
		SELECT device_id, mode, target_temperature
		FROM target_state
	`

	var states []thermostat.TargetState
	err := c.db.SelectContext(ctx, &states, query)
	if err != nil {
		return fmt.Errorf("error executing reportTargetStateMetrics query: %v", err)
	}

	for _, state := range states {
		if state.Mode != nil {
			metrics.SetThermostatMode(state.DeviceID, *state.Mode)
		}

		if state.TargetTemperature != nil {
			metrics.SetThermostatTargetTemperature(state.DeviceID, *state.TargetTemperature)
		}
	}

	return nil
}

func (c *Client) FetchTargetState(ctx context.Context, deviceID string) (*thermostat.TargetState, error) {
	query := `
		SELECT mode, target_temperature
		FROM target_state
		WHERE device_id = $1;
	`

	var data struct {
		Mode              sql.NullString `db:"mode"`
		TargetTemperature sql.NullInt32  `db:"target_temperature"`
	}
	err := c.db.GetContext(ctx, &data, query, deviceID)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("error executing FetchTargetState query: %v", err)
	}

	state := thermostat.TargetState{
		DeviceID: deviceID,
	}

	if data.Mode.Valid {
		modeValue := thermostat.Mode(data.Mode.String)
		state.Mode = &modeValue
	} else {
		state.Mode = &c.defaultMode
		err := c.updateMode(ctx, deviceID, c.defaultMode)
		if err != nil {
			return nil, fmt.Errorf("error setting default mode: %v", err)
		}
	}

	if data.TargetTemperature.Valid {
		targetTemperatureValue := int(data.TargetTemperature.Int32)
		state.TargetTemperature = &targetTemperatureValue
	} else {
		err := c.updateTargetTemperature(ctx, deviceID, c.defaultTargetTemperature)
		if err != nil {
			return nil, fmt.Errorf("error setting default target temperature: %v", err)
		}
		state.TargetTemperature = &c.defaultTargetTemperature
	}

	return &state, nil
}

func (c *Client) UpdateTargetState(ctx context.Context, state *thermostat.TargetState) (*thermostat.TargetState, error) {
	if state.Mode != nil {
		err := c.updateMode(ctx, state.DeviceID, *state.Mode)
		if err != nil {
			return nil, fmt.Errorf("error updating mode: %v", err)
		}
	}

	if state.TargetTemperature != nil {
		err := c.updateTargetTemperature(ctx, state.DeviceID, *state.TargetTemperature)
		if err != nil {
			return nil, fmt.Errorf("error updating target temperature: %v", err)
		}
	}

	return c.FetchTargetState(ctx, state.DeviceID)
}

func (c *Client) updateMode(ctx context.Context, deviceID string, mode thermostat.Mode) error {
	query := `
		INSERT INTO target_state (device_id, mode)
		VALUES ($1, $2)
		ON CONFLICT(device_id) DO UPDATE SET mode = $2;
	`

	_, err := c.db.ExecContext(ctx, query, deviceID, mode)
	if err != nil {
		return fmt.Errorf("error executing updateMode query: %v", err)
	}

	return nil
}

func (c *Client) updateTargetTemperature(ctx context.Context, deviceID string, targetTemperature int) error {
	query := `
		INSERT INTO target_state (device_id, target_temperature)
		VALUES ($1, $2)
		ON CONFLICT(device_id) DO UPDATE SET target_temperature = $2;
	`

	_, err := c.db.ExecContext(ctx, query, deviceID, targetTemperature)
	if err != nil {
		return fmt.Errorf("error executing updateTargetTemperature query: %v", err)
	}

	return nil
}
