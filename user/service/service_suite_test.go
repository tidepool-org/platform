package service_test

import (
	"testing"

	"github.com/Shopify/sarama"
	"github.com/onsi/ginkgo"

	kafkaTest "github.com/tidepool-org/platform/kafka/test"
	"github.com/tidepool-org/platform/test"
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
