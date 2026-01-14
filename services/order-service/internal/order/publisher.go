package order

import (
	"context"
	"encoding/json"
	"time"

	"uop-ms/pkg/events"

	"github.com/google/uuid"
)

type Publisher struct {
	producer *events.Producer
	topic    string
}

type OrderCreatedPayload struct {
	OrderID  string  `json:"orderId"`
	UserSub  string  `json:"userSub"`
	Total    float64 `json:"total"`
	Currency string  `json:"currency"`
}

func NewPublisher(producer *events.Producer, topic string) *Publisher {
	return &Publisher{producer: producer, topic: topic}
}

func (p *Publisher) PublishOrderCreated(ctx context.Context, traceID string, payload OrderCreatedPayload) error {
	env := events.Envelope[OrderCreatedPayload]{
		EventID:    uuid.NewString(),
		EventType:  events.EventTypeOrderCreated,
		OccurredAt: time.Now().UTC(),
		TraceID:    traceID,
		SchemaVer:  1,
		Producer:   "order-service",
		Payload:    payload,
	}

	b, err := json.Marshal(env)
	if err != nil {
		return err
	}

	// key = orderId so it hashes to same partition for same order
	return p.producer.Publish(ctx, payload.OrderID, b)
}
