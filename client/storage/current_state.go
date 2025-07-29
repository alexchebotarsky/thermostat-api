package storage

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/alexchebotarsky/thermostat-api/client"
	"github.com/alexchebotarsky/thermostat-api/model/thermostat"
)

func (c *Client) initCurrentStateTable(ctx context.Context) error {
	schema := `
		CREATE TABLE IF NOT EXISTS current_state (
			device_id TEXT PRIMARY KEY,
			timestamp DATETIME,
			operating_state TEXT,
			current_temperature REAL
		);
	`

	_, err := c.db.ExecContext(ctx, schema)
	if err != nil {
		return fmt.Errorf("error executing current state schema: %v", err)
	}

	return nil
}

func (c *Client) FetchCurrentState(ctx context.Context, deviceID string) (*thermostat.CurrentState, error) {
	query := `
		SELECT device_id, timestamp, operating_state, current_temperature
		FROM current_state
		WHERE device_id = $1;
	`

	var state thermostat.CurrentState
	err := c.db.GetContext(ctx, &state, query, deviceID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, &client.ErrNotFound{Err: err}
		} else {
			return nil, fmt.Errorf("error executing FetchCurrentState query: %v", err)
		}
	}

	return &state, nil
}

func (c *Client) UpdateCurrentState(ctx context.Context, state *thermostat.CurrentState) (*thermostat.CurrentState, error) {
	query := `
		INSERT INTO current_state (device_id, timestamp, operating_state, current_temperature)
		VALUES (:device_id, :timestamp, :operating_state, :current_temperature)
		ON CONFLICT(device_id) DO UPDATE SET
			timestamp = :timestamp,
			operating_state = :operating_state,
			current_temperature = :current_temperature;
	`

	_, err := c.db.NamedExecContext(ctx, query, state)
	if err != nil {
		return nil, fmt.Errorf("error executing UpdateCurrentState statement: %v", err)
	}

	return c.FetchCurrentState(ctx, state.DeviceID)
}
