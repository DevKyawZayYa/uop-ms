package main

import (
	"log"
	"uop-ms/services/product-service/internal/app/config"
	"uop-ms/services/product-service/internal/core"
	"uop-ms/services/product-service/internal/routes"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.Load()

	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	r.Use(core.ErrorHandler())

	routes.Register(r)

	addr := ":" + cfg.Port
	log.Println("product-service listening on", addr)
	if err := r.Run(addr); err != nil {
		log.Fatal(err)
	}
}
