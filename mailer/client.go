package mailer

import (
	"github.com/tidepool-org/go-common/clients"
	"github.com/tidepool-org/go-common/events"
)

func Client() (clients.MailerClient, error) {
	config := events.NewConfig()
	if err := config.LoadFromEnv(); err != nil {
		return nil, err
	}

	config.KafkaTopic = clients.MailerTopic
	config.EventSource = config.KafkaConsumerGroup
	// We are using a sync producer which requires setting the variables below
	config.SaramaConfig.Producer.Return.Errors = true
	config.SaramaConfig.Producer.Return.Successes = true

	return clients.NewMailerClient(config)
}
