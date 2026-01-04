package main

import (
	"log"
	"uop-ms/services/api-gateway/internal/app/config"
	"uop-ms/services/api-gateway/internal/auth"
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
	r.Use(core.RequestID())
	r.Use(core.ErrorHandler())

	// health
	routes.Register(r)

	// auth boundary
	r.Use(auth.Middleware(cfg))

	// proxy routes
	err := proxy.Register(r, cfg.ProductServiceURL, cfg.OrderServiceURL, func(c *gin.Context) string {
		if v, ok := c.Get(auth.HeaderUserSub); ok {
			if s, ok := v.(string); ok {
				return s
			}
		}
		return ""
	})
	if err != nil {
		log.Fatal(err)
	}

	addr := ":" + cfg.Port
	log.Println("api-gateway listening on", addr)
	if err := r.Run(addr); err != nil {
		log.Fatal(err)
	}
}
