package processor

import (
	"github.com/alexchebotarsky/thermofridge-api/processor/event"
	"github.com/alexchebotarsky/thermofridge-api/processor/handler"
	"github.com/alexchebotarsky/thermofridge-api/processor/middleware"
)

func (p *Processor) setupEvents() {
	p.use(middleware.Metrics)

	p.handle(event.Event{
		Topic:   "thermofridge/current-state",
		Handler: handler.CurrentState(p.Clients.Database),
	})
}
