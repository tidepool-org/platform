package base

import (
	"context"
	"time"

	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/work"
)

type ProcessingFactory func() (work.Processing, error)

type Processor struct {
	typ               string
	quantity          int
	frequency         time.Duration
	processingFactory ProcessingFactory
}

func NewProcessor(typ string, quantity int, frequency time.Duration, processingFactory ProcessingFactory) (*Processor, error) {
	if typ == "" {
		return nil, errors.New("type is missing")
	}
	if quantity <= 0 {
		return nil, errors.New("quantity is invalid")
	}
	if frequency <= 0 {
		return nil, errors.New("frequency is invalid")
	}
	if processingFactory == nil {
		return nil, errors.New("processing factory is missing")
	}
	return &Processor{
		typ:               typ,
		quantity:          quantity,
		frequency:         frequency,
		processingFactory: processingFactory,
	}, nil
}

func (p *Processor) Type() string {
	return p.typ
}

func (p *Processor) Quantity() int {
	return p.quantity
}

func (p *Processor) Frequency() time.Duration {
	return p.frequency
}

func (p *Processor) Process(ctx context.Context, wrk *work.Work, updater work.ProcessingUpdater) *work.ProcessResult {
	if processing, err := p.processingFactory(); err != nil {
		return NewProcessResultFailedFromError(err)
	} else {
		return processing.Process(ctx, wrk, updater)
	}
}
