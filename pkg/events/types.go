package events

import "time"

type Envelope[T any] struct {
	EventID    string    `json:"eventId"`
	EventType  string    `json:"eventType"`
	OccurredAt time.Time `json:"occurredAt"`
	TraceID    string    `json:"traceId"`
	SchemaVer  int       `json:"schemaVersion"`
	Producer   string    `json:"producer"`
	Payload    T         `json:"payload"`
}

const (
	EventTypeOrderCreated = "OrderCreated"
)
