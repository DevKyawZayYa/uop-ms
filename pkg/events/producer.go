package events

import (
	"context"
	"time"

	"github.com/segmentio/kafka-go"
)

type Producer struct {
	w *kafka.Writer
}

type ProducerConfig struct {
	Brokers []string
	Topic   string
}

func NewProducer(cfg ProducerConfig) *Producer {
	w := &kafka.Writer{
		Addr:         kafka.TCP(cfg.Brokers...),
		Topic:        cfg.Topic,
		Balancer:     &kafka.Hash{}, // stable partition by key
		BatchTimeout: 50 * time.Millisecond,
		RequiredAcks: kafka.RequireOne,
		Async:        false,
	}
	return &Producer{w: w}
}

func (p *Producer) Close() error {
	return p.w.Close()
}

func (p *Producer) Publish(ctx context.Context, key string, value []byte) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return p.w.WriteMessages(ctx, kafka.Message{
		Key:   []byte(key),
		Value: value,
		Time:  time.Now(),
	})
}
