package routes

import (
	"net/http"
	"uop-ms/services/product-service/internal/core"

	"github.com/gin-gonic/gin"
)

func Register(r *gin.Engine) {
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, core.APIResponse{
			Success: true,
			Data: gin.H{
				"status":  "ok",
				"service": "product-service"},
		})
	})
}
