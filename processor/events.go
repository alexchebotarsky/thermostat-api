package processor

import (
	"github.com/alexchebotarsky/thermostat-api/processor/event"
	"github.com/alexchebotarsky/thermostat-api/processor/handler"
	"github.com/alexchebotarsky/thermostat-api/processor/middleware"
)

func (p *Processor) setupEvents() {
	p.use(middleware.Metrics)

	p.handle(event.Event{
		Topic:   "thermostat/current-state",
		Handler: handler.CurrentState(p.Clients.Storage),
	})
}
