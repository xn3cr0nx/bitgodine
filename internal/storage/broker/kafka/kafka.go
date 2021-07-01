package broker

import (
	"context"
	"time"

	"github.com/segmentio/kafka-go"
)

// Kafka instance of Kafka writer
type Kafka struct {
	*kafka.Writer
	*kafka.Reader
}

// NewKafka initialize kafka broker
func NewKafka(kafkaBrokerUrls []string, topic string) (*Kafka, error) {
	w := &kafka.Writer{
		Addr:        kafka.TCP(kafkaBrokerUrls...),
		Topic:       topic,
		Balancer:    &kafka.LeastBytes{},
		Compression: kafka.Snappy,
	}

	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:   kafkaBrokerUrls,
		Topic:     topic,
		Partition: 0,
		MinBytes:  10e3, // 10KB
		MaxBytes:  10e6, // 10MB
	})

	return &Kafka{w, r}, nil
}

// Push writes message to kafka topic
func (k *Kafka) Push(ctx context.Context, key, value string) (err error) {
	message := kafka.Message{
		Key:   []byte(key),
		Value: []byte(value),
		Time:  time.Now(),
	}
	return k.WriteMessages(ctx, message)
}

// Pull reads a message from kafka topic
func (k *Kafka) Pull(ctx context.Context) (m kafka.Message, err error) {
	m, err = k.ReadMessage(ctx)
	return
}

// Close closes reader connection
func (k *Kafka) Close() (err error) {
	if err = k.Writer.Close(); err != nil {
		return
	}
	return k.Reader.Close()
}
