package proxy

import (
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/gin-gonic/gin"
)

func newReverseProxy(target string) (*httputil.ReverseProxy, error) {
	u, err := url.Parse(target)
	if err != nil {
		return nil, err
	}

	proxy := httputil.NewSingleHostReverseProxy(u)

	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)
		// keep original path + query
	}

	return proxy, nil
}

func Register(r *gin.Engine, productURL, orderURL string, getUserSub func(*gin.Context) string) error {
	productProxy, err := newReverseProxy(productURL)
	if err != nil {
		return err
	}

	orderProxy, err := newReverseProxy(orderURL)
	if err != nil {
		return err
	}

	v1 := r.Group("/api/v1")
	{
		// products
		v1.Any("/products", func(c *gin.Context) {
			userSub := getUserSub(c)
			if userSub != "" {
				c.Request.Header.Set("X-User-Sub", userSub)
			}
			productProxy.ServeHTTP(c.Writer, c.Request)
		})
		v1.Any("/products/*path", func(c *gin.Context) {
			userSub := getUserSub(c)
			if userSub != "" {
				c.Request.Header.Set("X-User-Sub", userSub)
			}
			productProxy.ServeHTTP(c.Writer, c.Request)
		})

		// orders
		v1.Any("/orders", func(c *gin.Context) {
			userSub := getUserSub(c)
			if userSub != "" {
				c.Request.Header.Set("X-User-Sub", userSub)
			}
			orderProxy.ServeHTTP(c.Writer, c.Request)
		})
		v1.Any("/orders/*path", func(c *gin.Context) {
			userSub := getUserSub(c)
			if userSub != "" {
				c.Request.Header.Set("X-User-Sub", userSub)
			}
			orderProxy.ServeHTTP(c.Writer, c.Request)
		})
	}

	return nil
}
