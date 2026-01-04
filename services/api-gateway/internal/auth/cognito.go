package auth

import (
	"net/http"
	"strings"
	"uop-ms/services/api-gateway/internal/app/config"
	"uop-ms/services/api-gateway/internal/core"

	"github.com/gin-gonic/gin"
)

func CognitoMiddleware(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		//allow public product GETs
		if c.Request.Method == http.MethodGet && strings.HasPrefix(c.Request.URL.Path, "/api/v1/products") {
			c.Next()
			return
		}

		authz := c.GetHeader("Authorization")
		if !strings.HasPrefix(authz, "Bearer") {
			c.Error(core.NewBadRequest("MISSING_TOKEN", "Authorization Bearer token required"))
			c.Abort()
			return
		}

		c.Error(core.NewInternal("COGNITO_NOT_READY", "Cognito validation not implemented yet"))
		c.Abort()
	}
}
