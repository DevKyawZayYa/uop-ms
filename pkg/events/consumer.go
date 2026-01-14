package events

import (
	"context"
	"log"
	"time"

	"github.com/segmentio/kafka-go"
)

type Consumer struct {
	r *kafka.Reader
}

type ConsumerConfig struct {
	Brokers []string
	Topic   string
	GroupID string
}

func NewConsumer(cfg ConsumerConfig) *Consumer {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        cfg.Brokers,
		Topic:          cfg.Topic,
		GroupID:        cfg.GroupID,
		MinBytes:       1e3,
		MaxBytes:       10e6,
		CommitInterval: time.Second, // at-least-once
	})
	return &Consumer{r: r}
}

func (c *Consumer) Close() error {
	return c.r.Close()
}

type MessageHandler func(ctx context.Context, key string, value []byte) error

func (c *Consumer) Run(ctx context.Context, handler MessageHandler) error {
	for {
		m, err := c.r.FetchMessage(ctx)
		if err != nil {
			// ctx canceled => normal shutdown
			return err
		}

		hCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
		hErr := handler(hCtx, string(m.Key), m.Value)
		cancel()

		if hErr != nil {
			// do not commit => message can be retried
			log.Printf("[kafka] handler error: %v", hErr)
			continue
		}

		if err := c.r.CommitMessages(ctx, m); err != nil {
			log.Printf("[kafka] commit error: %v", err)
		}
	}
}
