package service_test

import (
	"github.com/Shopify/sarama"
	"github.com/onsi/ginkgo"
	"testing"

	"github.com/tidepool-org/platform/test"
	kafkaTest "github.com/tidepool-org/platform/kafka/test"
)

func TestSuite(t *testing.T) {
	test.Test(t)
}

var broker *sarama.MockBroker

var _ = ginkgo.BeforeSuite(func() {
	broker = kafkaTest.NewMockBroker(&testing.T{})
})

var _ = ginkgo.AfterSuite(func() {
	broker.Close()
})
