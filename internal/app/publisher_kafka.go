package app

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/segmentio/kafka-go"
)

type kafkaPublisher struct {
	writer *kafka.Writer
}

func newKafkaPublisher(brokers []string, topic, clientID string) (*kafkaPublisher, error) {
	if len(brokers) == 0 {
		return nil, errors.New("kafka brokers are required")
	}
	if topic == "" {
		return nil, errors.New("kafka topic is required")
	}
	if clientID == "" {
		clientID = "helloworld"
	}
	transport := &kafka.Transport{
		ClientID:    clientID,
		DialTimeout: 10 * time.Second,
		IdleTimeout: 30 * time.Second,
	}
	writer := &kafka.Writer{
		Addr:                   kafka.TCP(brokers...),
		Topic:                  topic,
		Balancer:               &kafka.Hash{},
		AllowAutoTopicCreation: false,
		Transport:              transport,
	}
	return &kafkaPublisher{writer: writer}, nil
}

func (p *kafkaPublisher) Publish(ctx context.Context, item api) error {
	payload, err := json.Marshal(item)
	if err != nil {
		return fmt.Errorf("marshal kafka payload: %w", err)
	}
	return p.writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(item.ID),
		Value: payload,
		Time:  time.Now(),
	})
}

func (p *kafkaPublisher) Close() error {
	return p.writer.Close()
}

func configureContentPublisher() func() {
	brokers := parseKafkaBrokers(os.Getenv("KAFKA_BROKERS"))
	topic := strings.TrimSpace(os.Getenv("KAFKA_TOPIC"))
	cleanup := func() {}

	if len(brokers) == 0 || topic == "" {
		log.Print("helloworld: kafka publisher disabled (missing KAFKA_BROKERS or KAFKA_TOPIC)")
		return cleanup
	}

	clientID := strings.TrimSpace(os.Getenv("KAFKA_CLIENT_ID"))
	if clientID == "" {
		clientID = serviceName
	}

	publisher, err := newKafkaPublisher(brokers, topic, clientID)
	if err != nil {
		log.Printf("helloworld: unable to initialize kafka publisher: %v", err)
		return cleanup
	}
	setContentPublisher(publisher)
	log.Printf("helloworld: kafka publisher enabled (topic=%s, brokers=%s)", topic, strings.Join(brokers, ","))

	return func() {
		if err := publisher.Close(); err != nil {
			log.Printf("helloworld: error closing kafka publisher: %v", err)
		}
		resetContentPublisher()
	}
}

func parseKafkaBrokers(raw string) []string {
	parts := strings.Split(raw, ",")
	var brokers []string
	for _, part := range parts {
		broker := strings.TrimSpace(part)
		if broker == "" {
			continue
		}
		brokers = append(brokers, broker)
	}
	return brokers
}
