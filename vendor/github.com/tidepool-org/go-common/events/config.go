package events

import (
	"errors"
	"github.com/IBM/sarama"
	"github.com/kelseyhightower/envconfig"
)

const DeadLetterSuffix = "-dead-letters"

type CloudEventsConfig struct {
	EventSource           string   `envconfig:"CLOUD_EVENTS_SOURCE" required:"false"`
	KafkaBrokers          []string `envconfig:"KAFKA_BROKERS" required:"true"`
	KafkaConsumerGroup    string   `envconfig:"KAFKA_CONSUMER_GROUP" required:"false"`
	KafkaTopic            string   `envconfig:"KAFKA_TOPIC" default:"events"`
	KafkaDeadLettersTopic string   `envconfig:"KAFKA_DEAD_LETTERS_TOPIC"`
	KafkaTopicPrefix      string   `envconfig:"KAFKA_TOPIC_PREFIX" required:"true"`
	KafkaRequireSSL       bool     `envconfig:"KAFKA_REQUIRE_SSL" required:"true"`
	KafkaVersion          string   `envconfig:"KAFKA_VERSION" required:"true"`
	KafkaUsername         string   `envconfig:"KAFKA_USERNAME" required:"false"`
	KafkaPassword         string   `envconfig:"KAFKA_PASSWORD" required:"false"`
	SaramaConfig          *sarama.Config
}

func NewConfig() *CloudEventsConfig {
	cfg := &CloudEventsConfig{}
	cfg.SaramaConfig = sarama.NewConfig()
	cfg.SaramaConfig.Consumer.Offsets.Initial = sarama.OffsetOldest
	return cfg
}

func (k *CloudEventsConfig) LoadFromEnv() error {
	if err := envconfig.Process("", k); err != nil {
		return err
	}
	version, err := sarama.ParseKafkaVersion(k.KafkaVersion)
	if err != nil {
		return err
	}
	k.SaramaConfig.Version = version
	if k.KafkaRequireSSL {
		k.SaramaConfig.Net.TLS.Enable = true
		// Use the root CAs of the host
		k.SaramaConfig.Net.TLS.Config.RootCAs = nil
	}
	if k.KafkaUsername != "" && k.KafkaPassword != "" {
		k.SaramaConfig.Net.SASL.Enable = true
		k.SaramaConfig.Net.SASL.Mechanism = sarama.SASLTypeSCRAMSHA512
		k.SaramaConfig.Net.SASL.User = k.KafkaUsername
		k.SaramaConfig.Net.SASL.Password = k.KafkaPassword
		k.SaramaConfig.Net.SASL.SCRAMClientGeneratorFunc = func() sarama.SCRAMClient {
			return &XDGSCRAMClient{HashGeneratorFcn: SHA512}
		}
	}

	return nil
}

func (k *CloudEventsConfig) GetPrefixedTopic() string {
	return k.KafkaTopicPrefix + k.KafkaTopic
}

func (k *CloudEventsConfig) GetDeadLettersTopic() string {
	if k.KafkaDeadLettersTopic == "" {
		return k.KafkaDeadLettersTopic
	}
	return k.KafkaTopicPrefix + k.KafkaDeadLettersTopic
}

func (k *CloudEventsConfig) IsDeadLettersEnabled() bool {
	return k.GetDeadLettersTopic() != ""
}

func validateConsumerConfig(config *CloudEventsConfig) error {
	if config.KafkaConsumerGroup == "" {
		return errors.New("consumer group cannot be empty")
	}
	return nil
}
