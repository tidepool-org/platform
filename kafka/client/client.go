package kafka

import (
	"context"
	"log"
	"os"

	"github.com/Shopify/sarama"
	cloudevents "github.com/cloudevents/sdk-go"
	"github.com/cloudevents/sdk-go/protocol/kafka_sarama/v2"
	"github.com/google/uuid"
)

type Kafka interface {
	kafkaSender() *kafka_sarama.Sender
	kafkaClient(Sender *kafka_sarama.Sender) cloudevents.Client
	KafkaProducer(event string, newUser string)
}

func kafkaSender() *kafka_sarama.Sender {
	topics, _ := os.LookupEnv("KAFKA_TOPIC")
	broker, _ := os.LookupEnv("KAFKA_BROKERS")
	log.Println(broker)
	log.Println(topics)

	saramaConfig := sarama.NewConfig()
	saramaConfig.Version = sarama.V2_0_0_0

	sender, err := kafka_sarama.NewSender([]string{"kafka-kafka-bootstrap.kafka.svc.cluster.local:9092"}, saramaConfig, "marketo")
	if err != nil {
		log.Printf("failed to create protocol: %s", err.Error())
	}
	return sender
}
func kafkaClient(Sender *kafka_sarama.Sender) cloudevents.Client {
	c, err := cloudevents.NewClient(Sender, cloudevents.WithTimeNow(), cloudevents.WithUUIDs())
	if err != nil {
		log.Printf("failed to create client, %v", err)
	}
	return c
}

func (a *Kafka) KafkaProducer(event string, newUser string) {
	e := cloudevents.NewEvent()
	e.SetID(uuid.New().String())
	e.SetType(event)
	e.SetSource("github.com/tidepool-org/shoreline/user/marketo")
	_ = e.SetData(cloudevents.ApplicationJSON, map[string]interface{}{
		"user":  newUser,
		"event": event,
	})

	if result := a.Cloudevents.Send(
		// Set the producer message key
		kafka_sarama.WithMessageKey(context.Background(), sarama.StringEncoder(e.ID())),
		e,
	); cloudevents.IsUndelivered(result) {
		log.Println("failed to send message")
		a.Cloudevents.Send(kafka_sarama.WithMessageKey(context.Background(), sarama.StringEncoder(e.ID())), e)
	} else {
		log.Printf("sent: %s %s, accepted: %t", event, newUser, cloudevents.IsACK(result))
	}
}
