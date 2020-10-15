package test

import (
	"testing"

	"github.com/Shopify/sarama"
)

func NewMockBroker(t *testing.T) *sarama.MockBroker {
	mockBroker := sarama.NewMockBrokerAddr(t, 0, "localhost:9092")
	mockBroker.SetHandlerByMap(map[string]sarama.MockResponse{
		"MetadataRequest": sarama.NewMockMetadataResponse(t).
			SetBroker(mockBroker.Addr(), mockBroker.BrokerID()).
			SetLeader("test-events", 0, mockBroker.BrokerID()),
	})
	return mockBroker
}
