package service

import (
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/work"
)

type Logger interface{}

type WorkClient interface{}

type Coordinator struct {
	logger     Logger
	workClient WorkClient
}

func NewCoordinator(logger Logger, workClient WorkClient) (*Coordinator, error) {
	if logger == nil {
		return nil, errors.New("logger is missing")
	}
	if workClient == nil {
		return nil, errors.New("work client is missing")
	}

	return &Coordinator{
		logger:     logger,
		workClient: workClient,
	}, nil
}

func (c *Coordinator) RegisterProcessors(processors []work.Processor) error {
	for _, processor := range processors {
		if err := c.RegisterProcessor(processor); err != nil {
			return err
		}
	}
	return nil
}

func (c *Coordinator) RegisterProcessor(processor work.Processor) error {
	if processor == nil {
		return errors.New("processor is missing")
	}

	return nil
}

func (c *Coordinator) Start() {}

func (c *Coordinator) Stop() {}
