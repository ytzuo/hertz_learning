package worker

import (
	"context"
	"fmt"

	"Hertz/infra/mq"
)

type OrderConsumer struct{}

func NewOrderConsumer() *OrderConsumer {
	return &OrderConsumer{}
}

func (c *OrderConsumer) HandleOrderCreated(ctx context.Context, event mq.Event) error {
	fmt.Printf("worker order.created reserve workflow key=%s payload=%v\n", event.Key, event.Payload)
	return nil
}

func (c *OrderConsumer) HandleOrderPaid(ctx context.Context, event mq.Event) error {
	fmt.Printf("worker order.paid invoice workflow key=%s payload=%v\n", event.Key, event.Payload)
	fmt.Printf("worker order.paid notify workflow key=%s payload=%v\n", event.Key, event.Payload)
	return nil
}
