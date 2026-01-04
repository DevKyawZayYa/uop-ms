package auth

import (
	"net/http"
	"strings"
	"uop-ms/services/api-gateway/internal/app/config"
	"uop-ms/services/api-gateway/internal/core"

	"github.com/gin-gonic/gin"
)

const (
	HeaderDevUserSub = "X-Dev-User-Sub"
	HeaderUserSub    = "X-User-Sub"
)

func Middleware(cfg *config.Config) gin.HandlerFunc {
	switch strings.ToLower(cfg.AuthMode) {
	case "cognito":
		return CognitoMiddleware(cfg) // implemented next (stub for now)
	default:
		return DevMiddleware()
	}
}

func DevMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Public endpoints can skip auth:
		if c.Request.Method == http.MethodGet && strings.HasPrefix(c.Request.URL.Path, "/api/v1/products") {
			c.Next()
			return
		}

		userSub := strings.TrimSpace(c.GetHeader(HeaderDevUserSub))
		if userSub == "" {
			c.Error(core.NewBadRequest("MISSING_DEV_USER", "X-Dev-User-Sub header required in dev mode"))
			c.Abort()
			return
		}

		c.Set(HeaderUserSub, userSub)
		c.Next()
	}
}
