package config

import (
	"strings"
)

type KafkaConfig struct {
	Brokers []string
	Topic   string
}

func LoadKafkaConfig() KafkaConfig {
	return KafkaConfig{
		Brokers: splitCSV(getEnv("KAFKA_BOOTSTRAP_SERVERS", "localhost:29092")),
		Topic:   getEnv("KAFKA_ORDER_EVENTS_TOPIC", "uop.order.events.v1"),
	}
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
