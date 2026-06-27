package app

import (
	"crypto/subtle"
	"strings"

	"github.com/xDarkicex/nanite"
)

// AuthMiddleware returns a nanite middleware that validates Bearer tokens.
// If token is empty, the middleware is a no-op.
func AuthMiddleware(token string) func(*nanite.Context, func()) {
	if token == "" {
		return func(c *nanite.Context, next func()) { next() }
	}

	return func(c *nanite.Context, next func()) {
		auth := c.Request.Header.Get("Authorization")
		if !strings.HasPrefix(auth, "Bearer ") {
			c.JSON(401, map[string]string{"error": "missing Authorization header"})
			c.Abort()
			return
		}
		if subtle.ConstantTimeCompare([]byte(auth[7:]), []byte(token)) != 1 {
			c.JSON(401, map[string]string{"error": "invalid token"})
			c.Abort()
			return
		}
		next()
	}
}
