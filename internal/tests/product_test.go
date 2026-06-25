package tests

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/assanoff/servicekit/logger"

	"github.com/assanoff/service-kit-x/internal/app/server"
)

// TestProductCRUD exercises the product module end to end (auth disabled by
// default, so the group is open).
func TestProductCRUD(t *testing.T) {
	if testing.Short() {
		t.Skip("integration test requires docker")
	}

	ctx := context.Background()
	srv, _ := newTestServer(ctx, t)

	created := mustDo(t, srv, http.MethodPost, "/products",
		`{"name":"Keyboard","price":4999}`, http.StatusCreated)
	id, _ := created["id"].(string)
	if id == "" {
		t.Fatalf("expected an id, got %v", created)
	}
	if created["price"].(float64) != 4999 {
		t.Fatalf("price mismatch: %v", created)
	}

	got := mustDo(t, srv, http.MethodGet, "/products/"+id, "", http.StatusOK)
	if got["name"] != "Keyboard" {
		t.Fatalf("name mismatch: %v", got)
	}

	updated := mustDo(t, srv, http.MethodPut, "/products/"+id, `{"price":3999}`, http.StatusOK)
	if updated["price"].(float64) != 3999 {
		t.Fatalf("update did not apply: %v", updated)
	}
	if updated["name"] != "Keyboard" {
		t.Fatalf("update should preserve name: %v", updated)
	}

	assertStatus(t, doReq(t, srv, http.MethodDelete, "/products/"+id, ""), http.StatusNoContent)
	assertStatus(t, doReq(t, srv, http.MethodGet, "/products/"+id, ""), http.StatusNotFound)
}

// TestGroupVsRouteAuth contrasts the two authorization patterns under one server
// with auth enabled: product is guarded as a whole group (api.WithApp), so even
// reads need a token; user guards only writes (per-route), so user reads stay
// public.
func TestGroupVsRouteAuth(t *testing.T) {
	if testing.Short() {
		t.Skip("integration test requires docker")
	}

	ctx := context.Background()
	cfg := startPostgres(ctx, t)
	cfg.Auth.Enabled = true
	cfg.Auth.JWTSecret = "test-secret"
	cfg.Auth.RequiredRole = "admin"
	cfg.HTTP.RequestTimeout = 10 * time.Second

	log := logger.New(io.Discard, logger.Config{Service: "test", Level: logger.LevelError})
	slog.SetDefault(log.Slog())
	handler, err := server.Handler(ctx, cfg, log)
	if err != nil {
		t.Fatalf("build handler: %v", err)
	}
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)

	// product group is fully guarded (WithApp): a read without a token -> 401.
	assertStatus(t, doReq(t, srv, http.MethodGet, "/products", ""), http.StatusUnauthorized)

	// user reads stay public even with auth enabled (per-route auth on writes only).
	assertStatus(t, doReq(t, srv, http.MethodGet, "/users", ""), http.StatusOK)

	// user writes are guarded -> 401 without a token.
	assertStatus(t, doReq(t, srv, http.MethodPost, "/users",
		`{"email":"bob@example.com","name":"Bob"}`), http.StatusUnauthorized)

	// With a valid admin token, the product read succeeds.
	token := signToken(t, "test-secret", "admin")
	assertStatus(t, getWithToken(t, srv, "/products", token), http.StatusOK)
}

// getWithToken issues a GET carrying a bearer token.
func getWithToken(t *testing.T, srv *httptest.Server, path, token string) *http.Response {
	t.Helper()
	req, err := http.NewRequest(http.MethodGet, srv.URL+path, nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := srv.Client().Do(req)
	if err != nil {
		t.Fatalf("GET %s: %v", path, err)
	}
	return resp
}
