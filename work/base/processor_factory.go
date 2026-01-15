package base

import (
	"time"

	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/work"
)

type NewProcessorFunc func() (work.Processor, error)

type ProcessorFactory struct {
	typ              string
	quantity         int
	frequency        time.Duration
	newProcessorFunc NewProcessorFunc
}

func NewProcessorFactory(typ string, quantity int, frequency time.Duration, newProcessorFunc NewProcessorFunc) (*ProcessorFactory, error) {
	if typ == "" {
		return nil, errors.New("type is missing")
	}
	if quantity <= 0 {
		return nil, errors.New("quantity is invalid")
	}
	if frequency <= 0 {
		return nil, errors.New("frequency is invalid")
	}
	if newProcessorFunc == nil {
		return nil, errors.New("new processor func is missing")
	}
	return &ProcessorFactory{
		typ:              typ,
		quantity:         quantity,
		frequency:        frequency,
		newProcessorFunc: newProcessorFunc,
	}, nil
}

func (p *ProcessorFactory) Type() string {
	return p.typ
}

func (p *ProcessorFactory) Quantity() int {
	return p.quantity
}

func (p *ProcessorFactory) Frequency() time.Duration {
	return p.frequency
}

func (p *ProcessorFactory) New() (work.Processor, error) {
	return p.newProcessorFunc()
}
