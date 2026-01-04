package main

import (
	"log"
	"uop-ms/services/order-service/internal/app/config"
	"uop-ms/services/order-service/internal/app/db"
	"uop-ms/services/order-service/internal/core"
	"uop-ms/services/order-service/internal/order"
	"uop-ms/services/order-service/internal/routes"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.Load()

	gdb := db.Connect(cfg.MySQLDSN)

	if err := gdb.AutoMigrate(&order.Order{}, &order.OrderItem{}); err != nil {
		log.Fatal(err)
	}

	store := order.NewStore(gdb)
	svc := order.NewService(store)
	h := order.NewHandler(svc)

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
