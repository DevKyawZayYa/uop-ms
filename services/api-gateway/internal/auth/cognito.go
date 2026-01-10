package auth

import (
	"fmt"
	"net/http"
	"strings"
	"time"
	"uop-ms/services/api-gateway/internal/app/config"
	"uop-ms/services/api-gateway/internal/core"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

var jwks = newJWKSCache()

type cognitoClaims struct {
	Sub      string `json:"sub"`
	TokenUse string `json:"token_use"` // "access" or "id"
	Iss      string `json:"iss"`
	ClientID string `json:"client_id"` // access token claim

	jwt.RegisteredClaims
}

func CognitoMiddleware(cfg *config.Config) gin.HandlerFunc {
	jwksURL := fmt.Sprintf(
		"https://cognito-idp.%s.amazonaws.com/%s/.well-known/jwks.json",
		cfg.CognitoRegion,
		cfg.CognitoUserPoolID,
	)

	expectedIssuer := fmt.Sprintf(
		"https://cognito-idp.%s.amazonaws.com/%s",
		cfg.CognitoRegion,
		cfg.CognitoUserPoolID,
	)

	return func(c *gin.Context) {
		// Public endpoints: allow product GET list/detail without token
		if c.Request.Method == http.MethodGet && strings.HasPrefix(c.Request.URL.Path, "/api/v1/products") {
			c.Next()
			return
		}

		authz := c.GetHeader("Authorization")
		if !strings.HasPrefix(authz, "Bearer ") {
			c.Error(core.NewBadRequest("MISSING_TOKEN", "Authorization Bearer token required"))
			c.Abort()
			return
		}

		raw := strings.TrimSpace(strings.TrimPrefix(authz, "Bearer "))

		claims := &cognitoClaims{}
		tok, err := jwt.ParseWithClaims(raw, claims, func(token *jwt.Token) (any, error) {
			// only accept RS256
			if token.Method.Alg() != jwt.SigningMethodRS256.Alg() {
				return nil, fmt.Errorf("unexpected signing method: %s", token.Method.Alg())
			}

			kid, _ := token.Header["kid"].(string)
			if kid == "" {
				return nil, fmt.Errorf("missing kid")
			}

			// cache hit
			if k, ok := jwks.get(kid); ok {
				return k, nil
			}

			// refresh JWKS cache
			keys, ferr := fetchJWKS(jwksURL)
			if ferr != nil {
				return nil, ferr
			}
			jwks.set(keys, 30*time.Minute)

			if k, ok := keys[kid]; ok {
				return k, nil
			}
			return nil, fmt.Errorf("kid not found")
		})

		if err != nil || tok == nil || !tok.Valid {
			c.Error(core.NewBadRequest("INVALID_TOKEN", "Invalid or expired token"))
			c.Abort()
			return
		}

		// issuer validation
		if claims.Iss != expectedIssuer {
			c.Error(core.NewBadRequest("INVALID_ISSUER", "Invalid token issuer"))
			c.Abort()
			return
		}

		// ensure using access token
		if claims.TokenUse != "access" {
			c.Error(core.NewBadRequest("INVALID_TOKEN_USE", "Use Cognito access token"))
			c.Abort()
			return
		}

		// client_id validation (good practice)
		if cfg.CognitoClientID != "" && claims.ClientID != cfg.CognitoClientID {
			c.Error(core.NewBadRequest("INVALID_CLIENT", "Invalid client_id"))
			c.Abort()
			return
		}

		if claims.Sub == "" {
			c.Error(core.NewBadRequest("MISSING_SUB", "Token missing sub"))
			c.Abort()
			return
		}

		// set identity for proxy injection
		c.Set(HeaderUserSub, claims.Sub)
		c.Next()
	}
}
