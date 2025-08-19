package test

import (
	"log"
	"os"
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

func SetTestEnvironmentVariables() map[string]string {
	old := make(map[string]string)
	for k, v := range testConfig {
		if oldValue, exists := os.LookupEnv(k); exists {
			old[k] = oldValue
		}
		if err := os.Setenv(k, v); err != nil {
			log.Panicf("could not set env variable: %v", err)
		}
	}
	return old
}

func RestoreOldEnvironmentVariables(old map[string]string) {
	for k := range testConfig {
		var err error
		if old, exists := old[k]; exists {
			err = os.Setenv(k, old)
		} else {
			err = os.Unsetenv(k)
		}
		if err != nil {
			log.Panicf("could not reset env variable: %v", err)
		}
	}
}
