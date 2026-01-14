package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"uop-ms/pkg/events"
)

type OrderCreatedPayload struct {
	OrderID  string  `json:"orderId"`
	UserSub  string  `json:"userSub"`
	Total    float64 `json:"total"`
	Currency string  `json:"currency"`
}

func mustGetenv(key string) string {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		log.Fatalf("missing env: %s", key)
	}
	return v
}

func splitCSV(s string) []string {
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func main() {
	brokers := splitCSV(mustGetenv("KAFKA_BOOTSTRAP_SERVERS"))
	topic := mustGetenv("KAFKA_ORDER_EVENTS_TOPIC")
	group := mustGetenv("KAFKA_NOTIFICATION_GROUP_ID")

	consumer := events.NewConsumer(events.ConsumerConfig{
		Brokers: brokers,
		Topic:   topic,
		GroupID: group,
	})
	defer consumer.Close()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	log.Printf("[notification-service] consuming topic=%s group=%s brokers=%v", topic, group, brokers)

	err := consumer.Run(ctx, func(ctx context.Context, key string, value []byte) error {
		var env events.Envelope[OrderCreatedPayload]
		if err := json.Unmarshal(value, &env); err != nil {
			return err
		}

		// just log
		log.Printf("[OrderCreated] eventId=%s traceId=%s orderId=%s userSub=%s total=%.2f %s key=%s occurredAt=%s",
			env.EventID, env.TraceID, env.Payload.OrderID, env.Payload.UserSub, env.Payload.Total, env.Payload.Currency,
			key, env.OccurredAt.Format(time.RFC3339),
		)
		return nil
	})

	// When ctx canceled, Run returns error; treat cancel as normal
	if err != nil && ctx.Err() == nil {
		log.Fatalf("consumer stopped with error: %v", err)
	}

	log.Println("[notification-service] stopped")
}
