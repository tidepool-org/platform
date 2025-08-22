package test

import (
	"github.com/IBM/sarama"
	"github.com/onsi/ginkgo/v2"
)

func NewMockBroker(t ginkgo.FullGinkgoTInterface) *sarama.MockBroker {
	mockBroker := sarama.NewMockBroker(t, 1)

	// Set up broker to handle metadata and produce requests for multiple topics
	metadataResponse := sarama.NewMockMetadataResponse(t).
		SetBroker(mockBroker.Addr(), mockBroker.BrokerID()).
		SetLeader("test-emails", 0, mockBroker.BrokerID())

	produceResponse := sarama.NewMockProduceResponse(t).
		SetError("test-emails", 0, sarama.ErrNoError)

	apiVersionsResponse := sarama.NewMockApiVersionsResponse(t)

	mockBroker.SetHandlerByMap(map[string]sarama.MockResponse{
		"MetadataRequest":    metadataResponse,
		"ProduceRequest":     produceResponse,
		"ApiVersionsRequest": apiVersionsResponse,
	})

	return mockBroker
}

func SetKafkaConfig(t ginkgo.FullGinkgoTInterface, mockBroker *sarama.MockBroker) {
	t.Setenv("KAFKA_BROKERS", mockBroker.Addr())
	t.Setenv("KAFKA_TOPIC_PREFIX", "test-")
	t.Setenv("KAFKA_REQUIRE_SSL", "false")
	t.Setenv("KAFKA_VERSION", "2.4.0")
	t.Setenv("KAFKA_CONSUMER_GROUP", "platform-test")
}
