package mq

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type Event struct {
	Topic     string
	Key       string
	Payload   any
	CreatedAt time.Time
}

type Handler func(ctx context.Context, event Event)

type MemoryMQ struct {
	mu       sync.RWMutex
	handlers map[string][]Handler
}

// MemoryMQ 是进程内事件总线。
// 它用于演示生产者/消费者边界，不依赖 Kafka、RabbitMQ 或 RocketMQ。
func NewMemoryMQ() *MemoryMQ {
	return &MemoryMQ{
		handlers: map[string][]Handler{},
	}
}

func (mq *MemoryMQ) Subscribe(topic string, handler Handler) {
	mq.mu.Lock()
	defer mq.mu.Unlock()

	mq.handlers[topic] = append(mq.handlers[topic], handler)
}

func (mq *MemoryMQ) Publish(ctx context.Context, event Event) error {
	event.CreatedAt = time.Now()

	mq.mu.RLock()
	handlers := append([]Handler(nil), mq.handlers[event.Topic]...)
	mq.mu.RUnlock()

	fmt.Printf("mq publish topic=%s key=%s payload=%v\n", event.Topic, event.Key, event.Payload)
	for _, handler := range handlers {
		go handler(ctx, event)
	}

	return nil
}
