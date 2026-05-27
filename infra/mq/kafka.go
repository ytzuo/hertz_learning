package mq

import (
	"context"
	"encoding/json"
	"time"

	"github.com/segmentio/kafka-go"
)

type KafkaConfig struct {
	Brokers []string
}

type KafkaMQ struct {
	writer *kafka.Writer
}

func NewKafka(cfg KafkaConfig) *KafkaMQ {
	return &KafkaMQ{
		writer: &kafka.Writer{
			Addr:     kafka.TCP(cfg.Brokers...),
			Balancer: &kafka.LeastBytes{},
		},
	}
}

func (k *KafkaMQ) Publish(ctx context.Context, event Event) error {
	event.CreatedAt = time.Now()

	body, err := json.Marshal(event.Payload)
	if err != nil {
		return err
	}

	return k.writer.WriteMessages(ctx, kafka.Message{
		Topic: event.Topic,
		Key:   []byte(event.Key),
		Value: body,
		Time:  event.CreatedAt,
	})
}

func (k *KafkaMQ) Close() error {
	return k.writer.Close()
}
