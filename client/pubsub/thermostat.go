package pubsub

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/alexchebotarsky/thermostat-api/model/thermostat"
)

func (p *Client) PublishTargetState(ctx context.Context, state *thermostat.TargetState) error {
	payload, err := json.Marshal(&state)
	if err != nil {
		return fmt.Errorf("error marshalling target state: %v", err)
	}

	err = p.Publish(ctx, "thermostat/set/target-state", payload)
	if err != nil {
		return fmt.Errorf("error publishing target state: %v", err)
	}

	return nil
}
