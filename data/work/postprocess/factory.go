package postprocess

import (
	"strconv"
	"time"

	"github.com/tidepool-org/platform/clinics"
	"github.com/tidepool-org/platform/config"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/user"
	"github.com/tidepool-org/platform/work"
	workBase "github.com/tidepool-org/platform/work/base"
)

const (
	QuantityDefault = 60
	Frequency       = 5 * time.Second

	FailingRetryDuration        = 1 * time.Minute
	FailingRetryDurationJitter  = 0.1
	FailingRetryDurationMaximum = 1 * time.Hour
)

type (
	ClinicsClient = clinics.Client
	UserClient    = user.Client
)

type Config struct {
	Quantity int
}

func NewConfig() *Config {
	return &Config{
		Quantity: QuantityDefault,
	}
}

func (c *Config) Load(configReporter config.Reporter) error {
	if configReporter == nil {
		return errors.New("config reporter is missing")
	}
	if valueString, err := configReporter.Get("quantity"); err == nil {
		if value, parseErr := strconv.Atoi(valueString); parseErr != nil || value <= 0 {
			return errors.New("quantity is invalid")
		} else {
			c.Quantity = value
		}
	}
	return nil
}

func (c *Config) Validate() error {
	if c.Quantity <= 0 {
		return errors.New("quantity is invalid")
	}
	return nil
}

type Dependencies struct {
	workBase.Dependencies
	Summarizers
	ClinicsClient
	UserClient
}

func (d Dependencies) Validate() error {
	if err := d.Dependencies.Validate(); err != nil {
		return err
	}
	if d.Summarizers == nil {
		return errors.New("summarizers is missing")
	}
	if d.ClinicsClient == nil {
		return errors.New("clinics client is missing")
	}
	if d.UserClient == nil {
		return errors.New("user client is missing")
	}
	return nil
}

func NewProcessorFactory(config *Config, dependencies Dependencies) (*workBase.ProcessorFactory, error) {
	if config == nil {
		return nil, errors.New("config is missing")
	}
	if err := config.Validate(); err != nil {
		return nil, errors.Wrap(err, "config is invalid")
	}
	if err := dependencies.Validate(); err != nil {
		return nil, errors.Wrap(err, "dependencies is invalid")
	}
	processorFactory := func() (work.Processor, error) { return NewProcessor(dependencies) }
	return workBase.NewProcessorFactory(Type, config.Quantity, Frequency, processorFactory)
}
