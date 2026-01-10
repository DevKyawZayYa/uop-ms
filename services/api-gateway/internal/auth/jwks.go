package auth

import (
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"errors"
	"math/big"
	"net/http"
	"sync"
	"time"
)

type jwksKey struct {
	Kid string `json:"kid"`
	Kty string `json:"kty"`
	Alg string `json:"alg"`
	Use string `json:"use"`
	N   string `json:"n"`
	E   string `json:"e"`
}

type jwksResponse struct {
	Keys []jwksKey `json:"keys"`
}

type jwksCache struct {
	mu        sync.RWMutex
	keysByKid map[string]*rsa.PublicKey
	expiresAt time.Time
}

func newJWKSCache() *jwksCache {
	return &jwksCache{keysByKid: map[string]*rsa.PublicKey{}}
}

func (c *jwksCache) get(kid string) (*rsa.PublicKey, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if time.Now().After(c.expiresAt) {
		return nil, false
	}
	k, ok := c.keysByKid[kid]
	return k, ok
}

func (c *jwksCache) set(keys map[string]*rsa.PublicKey, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.keysByKid = keys
	c.expiresAt = time.Now().Add(ttl)
}

func fetchJWKS(jwksURL string) (map[string]*rsa.PublicKey, error) {
	resp, err := http.Get(jwksURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return nil, errors.New("failed to fetch jwks")
	}

	var jwks jwksResponse
	if err := json.NewDecoder(resp.Body).Decode(&jwks); err != nil {
		return nil, err
	}

	out := make(map[string]*rsa.PublicKey, len(jwks.Keys))
	for _, k := range jwks.Keys {
		pub, err := toRSAPublicKey(k.N, k.E)
		if err != nil {
			continue
		}
		out[k.Kid] = pub
	}
	return out, nil
}

func toRSAPublicKey(nB64, eB64 string) (*rsa.PublicKey, error) {
	nb, err := base64.RawURLEncoding.DecodeString(nB64)
	if err != nil {
		return nil, err
	}
	eb, err := base64.RawURLEncoding.DecodeString(eB64)
	if err != nil {
		return nil, err
	}

	n := new(big.Int).SetBytes(nb)

	e := 0
	for _, b := range eb {
		e = e<<8 + int(b)
	}

	return &rsa.PublicKey{N: n, E: e}, nil
}
