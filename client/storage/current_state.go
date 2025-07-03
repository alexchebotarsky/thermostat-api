package storage

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/alexchebotarsky/thermofridge-api/client"
	"github.com/alexchebotarsky/thermofridge-api/model/thermofridge"
)

func (c *Client) initCurrentStateTable(ctx context.Context) error {
	schema := `
		CREATE TABLE IF NOT EXISTS current_state (
			device_id TEXT PRIMARY KEY,
			operating_state TEXT,
			current_temperature REAL,
			timestamp DATETIME
		);
	`

	_, err := c.db.ExecContext(ctx, schema)
	if err != nil {
		return fmt.Errorf("error executing current state schema: %v", err)
	}

	return nil
}

func (c *Client) FetchCurrentState(ctx context.Context, deviceID string) (*thermofridge.CurrentState, error) {
	query := `
		SELECT device_id, operating_state, current_temperature, timestamp
		FROM current_state
		WHERE device_id = $1;
	`

	var state thermofridge.CurrentState
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

func (c *Client) UpdateCurrentState(ctx context.Context, state *thermofridge.CurrentState) (*thermofridge.CurrentState, error) {
	query := `
		INSERT INTO current_state (device_id, operating_state, current_temperature, timestamp)
		VALUES (:device_id, :operating_state, :current_temperature, :timestamp)
		ON CONFLICT(device_id) DO UPDATE SET
			operating_state = :operating_state,
			current_temperature = :current_temperature,
			timestamp = :timestamp;
	`

	_, err := c.db.NamedExecContext(ctx, query, state)
	if err != nil {
		return nil, fmt.Errorf("error executing UpdateCurrentState statement: %v", err)
	}

	return c.FetchCurrentState(ctx, state.DeviceID)
}
