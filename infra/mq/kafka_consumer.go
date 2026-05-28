package mq

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/segmentio/kafka-go"
)

type EventHandler func(ctx context.Context, event Event) error

type KafkaConsumerConfig struct {
	Brokers []string
	GroupID string
	Topics  []string
}

type KafkaConsumer struct {
	readers []*kafka.Reader
}

func NewKafkaConsumer(cfg KafkaConsumerConfig) *KafkaConsumer {
	readers := make([]*kafka.Reader, 0, len(cfg.Topics))
	for _, topic := range cfg.Topics {
		readers = append(readers, kafka.NewReader(kafka.ReaderConfig{
			Brokers: cfg.Brokers,
			GroupID: cfg.GroupID,
			Topic:   topic,
		}))
	}

	return &KafkaConsumer{readers: readers}
}

func (c *KafkaConsumer) Run(ctx context.Context, handlers map[string]EventHandler) error {
	errs := make(chan error, len(c.readers))
	var wg sync.WaitGroup

	for _, reader := range c.readers {
		wg.Add(1)
		go func(reader *kafka.Reader) {
			defer wg.Done()
			if err := c.consume(ctx, reader, handlers); err != nil {
				errs <- err
			}
		}(reader)
	}

	go func() {
		wg.Wait()
		close(errs)
	}()

	for err := range errs {
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *KafkaConsumer) consume(ctx context.Context, reader *kafka.Reader, handlers map[string]EventHandler) error {
	for {
		msg, err := reader.FetchMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return nil
			}
			return err
		}

		handler, ok := handlers[msg.Topic]
		if !ok {
			if err := reader.CommitMessages(ctx, msg); err != nil {
				return err
			}
			continue
		}

		event, err := decodeKafkaEvent(msg)
		if err != nil {
			return err
		}
		if err := handler(ctx, event); err != nil {
			return err
		}
		if err := reader.CommitMessages(ctx, msg); err != nil {
			return err
		}
	}
}

func (c *KafkaConsumer) Close() error {
	for _, reader := range c.readers {
		if err := reader.Close(); err != nil {
			return err
		}
	}
	return nil
}

func decodeKafkaEvent(msg kafka.Message) (Event, error) {
	var payload any
	if len(msg.Value) > 0 {
		if err := json.Unmarshal(msg.Value, &payload); err != nil {
			return Event{}, fmt.Errorf("decode kafka payload topic=%s key=%s: %w", msg.Topic, string(msg.Key), err)
		}
	}

	createdAt := msg.Time
	if createdAt.IsZero() {
		createdAt = time.Now()
	}

	return Event{
		Topic:     msg.Topic,
		Key:       string(msg.Key),
		Payload:   payload,
		CreatedAt: createdAt,
	}, nil
}
