package event

import "context"

type Event struct {
	Topic       string
	Handler     Handler
	Middlewares []Middleware
}

type Handler = func(ctx context.Context, payload []byte) error

type Middleware func(topic string, next Handler) Handler
