package proxy

import (
	"net/http"
	"uop-ms/services/api-gateway/internal/core"

	"github.com/gin-gonic/gin"
)

func RegisterPlaceHolders(r *gin.Engine) {
	v1 := r.Group("/api/v1")
	{
		v1.Any("/products/*path", func(c *gin.Context) {
			c.Error(core.NewInternal("NOT_IMPLEMENTED", "Gateway proxy not implemented yet!"))
		})
		v1.Any("/orders/*path", func(c *gin.Context) {
			c.Error(core.NewInternal("NOT_IMPLEMENTED", "Gateway proxy not implemented yet"))
		})
	}

	r.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, core.APIResponse{
			Success: false,
			Error: &core.APIError{
				Code:    "ROUTE_NOT_FOUND",
				Message: "Route not found",
			},
		})
	})
}
