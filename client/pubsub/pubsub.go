package pubsub

import (
	"context"
	"fmt"
	"log/slog"
	"net/url"

	"github.com/eclipse/paho.golang/autopaho"
	"github.com/eclipse/paho.golang/paho"
)

type Client struct {
	clientID      string
	qos           byte
	subscriptions map[string]func(ctx context.Context, payload []byte) error

	connManager *autopaho.ConnectionManager
}

func New(ctx context.Context, host string, port uint16, clientID string, qos byte) (*Client, error) {
	var p Client
	var err error

	p.clientID = clientID
	p.qos = qos
	p.subscriptions = make(map[string]func(ctx context.Context, payload []byte) error)

	brokerURL, err := url.Parse(fmt.Sprintf("mqtt://%s:%d", host, port))
	if err != nil {
		return nil, fmt.Errorf("error parsing broker URL: %v", err)
	}

	cfg := autopaho.ClientConfig{
		ServerUrls:            []*url.URL{brokerURL},
		SessionExpiryInterval: 10 * 60, // 10 minutes for reconnection
		OnConnectError:        p.handleConnectError,
		ClientConfig: paho.ClientConfig{
			ClientID: p.clientID,
			OnPublishReceived: []func(paho.PublishReceived) (bool, error){
				p.handleMessage,
			},
		},
	}

	p.connManager, err = autopaho.NewConnection(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("error creating pubsub connection: %v", err)
	}

	err = p.connManager.AwaitConnection(ctx)
	if err != nil {
		return nil, fmt.Errorf("error awaiting pubsub connection: %v", err)
	}

	return &p, nil
}

func (p *Client) Close(ctx context.Context) error {
	err := p.connManager.Disconnect(ctx)
	if err != nil {
		return fmt.Errorf("error disconnecting: %v", err)
	}

	return nil
}

func (p *Client) Publish(ctx context.Context, topic string, payload []byte) error {
	_, err := p.connManager.Publish(ctx, &paho.Publish{
		Topic:   topic,
		Payload: payload,
		QoS:     p.qos,
	})
	if err != nil {
		return fmt.Errorf("error publishing message: %v", err)
	}

	return nil
}

func (p *Client) Subscribe(ctx context.Context, topic string, handler func(ctx context.Context, payload []byte) error) error {
	p.subscriptions[topic] = handler

	_, err := p.connManager.Subscribe(ctx, &paho.Subscribe{
		Subscriptions: []paho.SubscribeOptions{
			{Topic: topic, QoS: p.qos},
		},
	})
	if err != nil {
		return fmt.Errorf("error subscribing to topic: %v", err)
	}

	return nil
}

func (p *Client) handleMessage(message paho.PublishReceived) (bool, error) {
	for topic, handler := range p.subscriptions {
		if message.Packet.Topic == topic {
			err := handler(context.Background(), message.Packet.Payload)
			if err != nil {
				slog.Error(fmt.Sprintf("Error handling message for topic %s: %v", topic, err))
				return true, err
			}
		}
	}

	return true, nil
}

func (p *Client) handleConnectError(err error) {
	slog.Error(fmt.Sprintf("error with pubsub connection: %v", err))
}
