package main

import (
	"log"
	"uop-ms/services/api-gateway/internal/app/config"
	"uop-ms/services/api-gateway/internal/core"
	"uop-ms/services/api-gateway/internal/proxy"
	"uop-ms/services/api-gateway/internal/routes"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.Load()

	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	r.Use(core.ErrorHandler())

	routes.Register(r)
	proxy.RegisterPlaceHolders(r)

	addr := ":" + cfg.Port
	log.Println("api-gateway listening on", addr)
	if err := r.Run(addr); err != nil {
		log.Fatal(err)
	}
}
