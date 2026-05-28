package bootstrap

import (
	"context"
	"fmt"
	"strings"

	workerbiz "Hertz/biz/worker"
	"Hertz/config"
	"Hertz/infra/mq"
)

type Worker struct {
	cfg config.Config
}

func NewWorker() *Worker {
	return &Worker{cfg: config.Load()}
}

func (w *Worker) Run(ctx context.Context) error {
	switch strings.ToLower(w.cfg.Infra.Adapter) {
	case "real":
		return w.runKafka(ctx)
	case "memory", "":
		fmt.Println("worker memory mode has no cross-process queue; API process consumes MemoryMQ events in-process")
		<-ctx.Done()
		return nil
	default:
		return fmt.Errorf("unsupported APP_ADAPTER %q", w.cfg.Infra.Adapter)
	}
}

func (w *Worker) runKafka(ctx context.Context) error {
	consumer := mq.NewKafkaConsumer(mq.KafkaConsumerConfig{
		Brokers: w.cfg.Infra.MQ.Brokers,
		GroupID: w.cfg.Infra.MQ.GroupID,
		Topics:  w.cfg.Infra.MQ.Topics,
	})
	defer consumer.Close()

	orderConsumer := workerbiz.NewOrderConsumer()
	return consumer.Run(ctx, map[string]mq.EventHandler{
		"order.created": orderConsumer.HandleOrderCreated,
		"order.paid":    orderConsumer.HandleOrderPaid,
	})
}
