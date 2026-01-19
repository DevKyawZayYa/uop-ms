package main

import (
	"context"
	"log"

	"github.com/gin-gonic/gin"

	"uop-ms/pkg/events"
	"uop-ms/pkg/redisx"

	"uop-ms/services/order-service/internal/app/config"
	"uop-ms/services/order-service/internal/app/db"
	"uop-ms/services/order-service/internal/core"
	"uop-ms/services/order-service/internal/order"
	"uop-ms/services/order-service/internal/routes"
)

func main() {
	// Existing config (MySQL + Port)
	cfg := config.Load()

	// Kafka config
	kCfg := config.LoadKafkaConfig()

	// DB
	gdb := db.Connect(cfg.MySQLDSN)

	if err := gdb.AutoMigrate(&order.Order{}, &order.OrderItem{}); err != nil {
		log.Fatal(err)
	}

	//Redis
	rCfg := redisx.LoadConfig()
	redisClient := redisx.New(rCfg)
	defer func() {
		_ = redisClient.Close()
	}()

	if err := redisClient.Ping(context.Background()); err != nil {
		log.Fatal("redis ping failed:", err)
	}

	// Kafka producer
	kProducer := events.NewProducer(events.ProducerConfig{
		Brokers: kCfg.Brokers,
		Topic:   kCfg.Topic,
	})
	defer func() {
		_ = kProducer.Close()
	}()

	// Kafka Publisher
	publisher := order.NewPublisher(kProducer, kCfg.Topic)

	// Existing DI, but Service now needs publisher Kafka
	store := order.NewStore(gdb)
	svc := order.NewService(store, publisher, redisClient.Raw())
	h := order.NewHandler(svc)

	// HTTP
	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	r.Use(core.ErrorHandler())

	routes.Register(r)
	h.Register(r)

	addr := ":" + cfg.Port
	log.Println("order-service listening on", addr)
	if err := r.Run(addr); err != nil {
		log.Fatal(err)
	}
}
