package storage

import (
	"context"
	"fmt"

	"github.com/alexchebotarsky/thermostat-api/model/thermostat"
	"github.com/jmoiron/sqlx"

	// sqlite driver
	_ "modernc.org/sqlite"
)

type Client struct {
	defaultMode              thermostat.Mode
	defaultTargetTemperature int

	db *sqlx.DB
}

func New(ctx context.Context, path string, defaultMode thermostat.Mode, defaultTargetTemperature int) (*Client, error) {
	var c Client
	var err error

	c.defaultMode = defaultMode
	c.defaultTargetTemperature = defaultTargetTemperature

	c.db, err = sqlx.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("error opening database: %v", err)
	}

	err = c.initTargetStateTable(ctx)
	if err != nil {
		return nil, fmt.Errorf("error initializing target state table: %v", err)
	}

	err = c.initCurrentStateTable(ctx)
	if err != nil {
		return nil, fmt.Errorf("error initializing current state table: %v", err)
	}

	return &c, nil
}

func (c *Client) Close() error {
	return c.db.Close()
}
