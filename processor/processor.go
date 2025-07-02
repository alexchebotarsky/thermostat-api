package processor

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/alexchebotarsky/thermofridge-api/processor/event"
	"github.com/alexchebotarsky/thermofridge-api/processor/handler"
)

type Processor struct {
	Events      []event.Event
	Middlewares []event.Middleware
	Clients     Clients
}

type Clients struct {
	PubSub  PubSubClient
	Storage StorageClient
}

type PubSubClient interface {
	Subscribe(ctx context.Context, topic string, handler event.Handler) error
}

type StorageClient interface {
	handler.CurrentStateUpdater
}

func New(clients Clients) *Processor {
	var p Processor

	p.Clients = clients

	p.setupEvents()

	return &p
}

func (p *Processor) Start(ctx context.Context, errc chan<- error) {
	for _, e := range p.Events {
		// Gather global processor middlewares and event specific middlewares
		middlewares := make([]event.Middleware, 0, len(p.Middlewares)+len(e.Middlewares))
		middlewares = append(middlewares, p.Middlewares...)
		middlewares = append(middlewares, e.Middlewares...)

		// Apply relevant middlewares before listening to the event
		for _, middleware := range middlewares {
			e.Handler = middleware(e.Topic, e.Handler)
		}

		err := p.Clients.PubSub.Subscribe(ctx, e.Topic, e.Handler)
		if err != nil {
			errc <- fmt.Errorf("error subscribing to topic %s: %v", e.Topic, err)
			return
		}
	}

	slog.Info(fmt.Sprintf("PubSub event processor listening to %d events", len(p.Events)))
}

func (p *Processor) Stop(ctx context.Context) error {
	return nil
}

func (p *Processor) handle(e event.Event) {
	p.Events = append(p.Events, e)
}

func (p *Processor) use(middlewares ...event.Middleware) {
	p.Middlewares = append(p.Middlewares, middlewares...)
}
