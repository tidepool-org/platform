package kafkasender

import (
	"context"
	"log"
	"os"

	"github.com/Shopify/sarama"
	"github.com/cloudevents/sdk-go/protocol/kafka_sarama/v2"
	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/google/uuid"
)

//CloudEventsClient is the method signature for Kafka message
type CloudEventsClient interface {
	KafkaMessage(event string, userID string, email string, role []string)
}

//Kafka struct containing the kafka topic and broker
type Kafka struct {
	topic  string
	broker string
}

//NewKafka creates a kafka struct containing the kafka topic and broker
func (k *Kafka) NewKafka() *Kafka {
	prefix, _ := os.LookupEnv("KAFKA_PREFIX")
	topic, _ := os.LookupEnv("KAFKA_TOPIC")
	topicWithPrefix := prefix + topic
	broker, _ := os.LookupEnv("KAFKA_BROKERS")
	log.Println(broker)
	log.Println(topic)

	return &Kafka{
		topicWithPrefix,
		broker,
	}
}

// Initialize COME BACK AND REVIEW THIS
var Initialize CloudEventsClient = &Kafka{}

// KafkaSender sends message to correct topic and broker
func (k *Kafka) KafkaSender() *kafka_sarama.Sender {
	saramaConfig := sarama.NewConfig()
	saramaConfig.Version = sarama.V2_0_0_0
	log.Printf("Broker: %v Topic: %v", k.broker, k.topic)
	log.Printf("New Broker: %v New Topic: %v", k.NewKafka().broker, k.NewKafka().topic)

	sender, err := kafka_sarama.NewSender([]string{k.NewKafka().broker}, saramaConfig, k.NewKafka().topic)
	if err != nil {
		log.Printf("failed to create protocol: %s", err.Error())
	}
	return sender
}

// KafkaClient builds kafka client
func (k *Kafka) KafkaClient(Sender *kafka_sarama.Sender) cloudevents.Client {
	c, err := cloudevents.NewClient(Sender, cloudevents.WithTimeNow(), cloudevents.WithUUIDs())
	if err != nil {
		log.Printf("failed to create client, %v", err)
	}
	return c
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
		"role":  role[0],
		"event": event,
	})

	kafkaSender := k.KafkaSender()
	defer kafkaSender.Close(context.Background())

	if result := k.KafkaClient(kafkaSender).Send(
		// Set the producer message key
		kafka_sarama.WithMessageKey(context.Background(), sarama.StringEncoder(e.ID())),
		e,
	); cloudevents.IsUndelivered(result) {
		log.Println("failed to send message")
		k.KafkaClient(kafkaSender).Send(kafka_sarama.WithMessageKey(context.Background(), sarama.StringEncoder(e.ID())), e)
	} else {
		log.Printf("sent: %s %s %s %s, accepted: %t", event, userID, email, role, cloudevents.IsACK(result))
	}
}
