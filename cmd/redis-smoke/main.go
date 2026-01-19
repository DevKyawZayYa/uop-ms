package main

import (
	"context"
	"log"

	"uop-ms/pkg/redisx"
)

func main() {
	cfg := redisx.LoadConfig()
	rc := redisx.New(cfg)
	defer rc.Close()

	if err := rc.Ping(context.Background()); err != nil {
		log.Fatal("redis ping failed:", err)
	}

	log.Println("redis ping ok")
}
