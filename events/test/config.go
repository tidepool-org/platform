package test

import (
	"github.com/onsi/ginkgo/v2"
)

var testConfig = map[string]string{
	"CLOUD_EVENTS_SOURCE":  "platform-test",
	"KAFKA_BROKERS":        "localhost:9092",
	"KAFKA_CONSUMER_GROUP": "platform-test",
	"KAFKA_TOPIC":          "events",
	"KAFKA_TOPIC_PREFIX":   "test-",
	"KAFKA_REQUIRE_SSL":    "false",
	"KAFKA_VERSION":        "2.4.0",
}

func SetTestEnvironmentVariables(t ginkgo.FullGinkgoTInterface) {
	for k, v := range testConfig {
		t.Setenv(k, v)
	}
}
