package app

import (
	"context"
	"crypto/tls"
	"crypto/x509"
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

func newKafkaPublisher(brokers []string, topic, clientID string, tlsConfig *tls.Config) (*kafkaPublisher, error) {
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
		TLS:         tlsConfig,
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

	var tlsConfig *tls.Config
	if strings.EqualFold(strings.TrimSpace(os.Getenv("KAFKA_TLS_ENABLED")), "true") {
		var err error
		tlsConfig, err = loadTLSConfig(
			strings.TrimSpace(os.Getenv("KAFKA_TLS_CA_FILE")),
			strings.TrimSpace(os.Getenv("KAFKA_TLS_CERT_FILE")),
			strings.TrimSpace(os.Getenv("KAFKA_TLS_KEY_FILE")),
		)
		if err != nil {
			log.Printf("helloworld: kafka TLS configuration invalid: %v", err)
			return cleanup
		}
	}

	publisher, err := newKafkaPublisher(brokers, topic, clientID, tlsConfig)
	if err != nil {
		log.Printf("helloworld: unable to initialize kafka publisher: %v", err)
		return cleanup
	}
	setContentPublisher(publisher)
	log.Printf("helloworld: kafka publisher enabled (topic=%s, brokers=%s, tls=%t)", topic, strings.Join(brokers, ","), tlsConfig != nil)

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

func loadTLSConfig(caFile, certFile, keyFile string) (*tls.Config, error) {
	switch {
	case caFile == "":
		return nil, errors.New("KAFKA_TLS_CA_FILE is required when TLS is enabled")
	case certFile == "":
		return nil, errors.New("KAFKA_TLS_CERT_FILE is required when TLS is enabled")
	case keyFile == "":
		return nil, errors.New("KAFKA_TLS_KEY_FILE is required when TLS is enabled")
	}

	clientCert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, fmt.Errorf("load client certificate: %w", err)
	}

	caPEM, err := os.ReadFile(caFile)
	if err != nil {
		return nil, fmt.Errorf("read CA file: %w", err)
	}
	caPool := x509.NewCertPool()
	if ok := caPool.AppendCertsFromPEM(caPEM); !ok {
		return nil, errors.New("failed to parse CA certificate")
	}

	return &tls.Config{
		MinVersion:   tls.VersionTLS12,
		RootCAs:      caPool,
		Certificates: []tls.Certificate{clientCert},
	}, nil
}
