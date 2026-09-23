package eventhub

import (
	"github.com/tidepool-org/platform/log"
)

type ServerSessionTokenProvider any

type ProviderSessionClient any

type DataSourceClient any

type DataSetClient any

type DataRawClient any

type WorkClient any

type ConsumerDependencies struct {
	Logger                     log.Logger
	ServerSessionTokenProvider ServerSessionTokenProvider
	ProviderSessionClient      ProviderSessionClient
	DataSourceClient           DataSourceClient
	DataSetClient              DataSetClient
	DataRawClient              DataRawClient
	WorkClient                 WorkClient
}

type Consumer struct{}

func NewConsumer(consumerDependencies ConsumerDependencies) (*Consumer, error) {
	return nil, nil
}

func (c *Consumer) Run() error {
	return nil
}

func (c *Consumer) Stop() {}
