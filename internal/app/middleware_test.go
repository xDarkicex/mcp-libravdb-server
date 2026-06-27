package app

import (
	"net/http/httptest"
	"testing"

	"github.com/xDarkicex/nanite"
)

func TestAuthMiddleware_NoToken_Passes(t *testing.T) {
	mw := AuthMiddleware("")
	r := nanite.New()
	r.Use(mw)
	r.Get("/test", func(c *nanite.Context) { c.String(200, "ok") })

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestAuthMiddleware_NoToken_Blocks(t *testing.T) {
	mw := AuthMiddleware("secret")
	r := nanite.New()
	r.Use(mw)
	r.Get("/test", func(c *nanite.Context) { c.String(200, "ok") })

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 401 {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestAuthMiddleware_WrongToken(t *testing.T) {
	mw := AuthMiddleware("secret")
	r := nanite.New()
	r.Use(mw)
	r.Get("/test", func(c *nanite.Context) { c.String(200, "ok") })

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer wrong")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 401 {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestAuthMiddleware_CorrectToken(t *testing.T) {
	mw := AuthMiddleware("secret")
	r := nanite.New()
	r.Use(mw)
	r.Get("/test", func(c *nanite.Context) { c.String(200, "ok") })

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer secret")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("expected 200, got %d — body=%q", w.Code, w.Body.String())
	}
}
