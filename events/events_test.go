package events

import (
	"testing"
	"time"
)

func configureTestKafka(t *testing.T) {
	t.Helper()
	for key, value := range map[string]string{
		"KAFKA_BROKERS": "127.0.0.1:1", "KAFKA_TOPIC_PREFIX": "test-", "KAFKA_TOPIC": "events",
		"KAFKA_CONSUMER_GROUP": "shared-environment", "KAFKA_REQUIRE_SSL": "false", "KAFKA_VERSION": "2.4.0",
	} {
		t.Setenv(key, value)
	}
}

func TestIndependentConsumerGroups(t *testing.T) {
	configureTestKafka(t)
	for _, group := range []string{"existing-auth", "existing-blob", "existing-data"} {
		runner := NewRunner(nil, WithConsumerGroup(group)).(*runner)
		config, err := runner.loadConfig()
		if err != nil {
			t.Fatal(err)
		}
		if config.KafkaConsumerGroup != group {
			t.Fatalf("group %s became %s", group, config.KafkaConsumerGroup)
		}
		if config.GetPrefixedTopic() != "test-events" {
			t.Fatal("topic configuration was lost")
		}
	}
	config, err := NewRunner(nil).(*runner).loadConfig()
	if err != nil {
		t.Fatal(err)
	}
	if config.KafkaConsumerGroup != "shared-environment" {
		t.Fatal("legacy group fallback was lost")
	}
}

func TestTerminateBeforeRunDoesNotStartConsumer(t *testing.T) {
	configureTestKafka(t)
	runner := NewRunner(nil)
	if err := runner.Initialize(); err != nil {
		t.Fatal(err)
	}
	if err := runner.Terminate(); err != nil {
		t.Fatal(err)
	}
	done := make(chan error, 1)
	go func() { done <- runner.Run() }()
	select {
	case err := <-done:
		if err == nil {
			t.Fatal("stopped consumer returned success")
		}
	case <-time.After(time.Second):
		t.Fatal("stopped consumer attempted to connect to Kafka")
	}
}
