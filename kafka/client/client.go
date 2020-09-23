package kafkasender

import (
	"context"
	"log"

	"github.com/Shopify/sarama"
	"github.com/cloudevents/sdk-go/protocol/kafka_sarama/v2"
	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/google/uuid"
	"github.com/kelseyhightower/envconfig"
)

//CloudEventsClient is the method signature for Kafka message
type CloudEventsClient interface {
	KafkaMessage(event string, userID string, email string, role []string)
}

//Kafka struct containing the kafka topic and broker
type Kafka struct {
	Prefix     string `envconfig:"KAFKA_PREFIX" required:"false"`
	BaseTopic  string `envconfig:"KAFKA_TOPIC" required:"false"`
	FinalTopic string
	Broker     string `envconfig:"KAFKA_BROKERS" required:"false"`
}

//NewServiceConfigFromEnv creates a kafka struct containing the kafka topic and broker
func NewServiceConfigFromEnv() (*Kafka, error) {
	var config Kafka
	err := envconfig.Process("", &config)
	config.FinalTopic = config.Prefix + config.BaseTopic
	return &config, err
}

// KafkaSender sends message to correct topic and broker
func (k *Kafka) KafkaSender() (*kafka_sarama.Sender, error) {
	saramaConfig := sarama.NewConfig()
	saramaConfig.Version = sarama.V2_0_0_0
	log.Printf("Broker: %v Topic: %v", k.Broker, k.FinalTopic)

	sender, err := kafka_sarama.NewSender([]string{k.Broker}, saramaConfig, k.FinalTopic)
	return sender, err
}

// KafkaClient builds kafka client
func (k *Kafka) KafkaClient(Sender *kafka_sarama.Sender) (cloudevents.Client, error) {
	c, err := cloudevents.NewClient(Sender, cloudevents.WithTimeNow(), cloudevents.WithUUIDs())
	return c, err
}

// KafkaMessage produces kafka message
func (k *Kafka) KafkaMessage(event string, userID string, email string, role []string) {
	e := cloudevents.NewEvent()
	e.SetID(uuid.New().String())
	e.SetType(event)
	e.SetSource("github.com/tidepool-org/platform/kafka/client")
	_ = e.SetData(cloudevents.ApplicationJSON, map[string]interface{}{
		"user":  userID,
		"email": email,
		"role":  role,
		"event": event,
	})

	kafkaSender, err := k.KafkaSender()
	if err != nil {
		log.Printf("failed to create client, %v", err)
	}
	defer kafkaSender.Close(context.Background())

	kafkaClient, err := k.KafkaClient(kafkaSender)
	if err != nil {
		log.Printf("failed to create protocol: %s", err.Error())
	}

	if result := kafkaClient.Send(
		// Set the producer message key
		kafka_sarama.WithMessageKey(context.Background(), sarama.StringEncoder(e.ID())),
		e,
	); cloudevents.IsUndelivered(result) {
		log.Println("failed to send message")
		kafkaClient.Send(kafka_sarama.WithMessageKey(context.Background(), sarama.StringEncoder(e.ID())), e)
	} else {
		log.Printf("sent: %s %s %v %v, accepted: %t", event, userID, email, role, cloudevents.IsACK(result))
	}
}
