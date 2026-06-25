package tests

import (
	"context"
	"net/http"
	"testing"
)

// TestUserCRUD exercises the user module end to end (auth disabled by default in
// newTestServer, so writes are open). It mirrors the widget CRUD test against the
// /users endpoints and the query.Result list envelope.
func TestUserCRUD(t *testing.T) {
	if testing.Short() {
		t.Skip("integration test requires docker")
	}

	ctx := context.Background()
	srv, _ := newTestServer(ctx, t)

	// Create.
	created := mustDo(t, srv, http.MethodPost, "/users",
		`{"email":"alice@example.com","name":"Alice"}`, http.StatusCreated)
	id, _ := created["id"].(string)
	if id == "" {
		t.Fatalf("expected an id, got %v", created)
	}
	if created["email"] != "alice@example.com" {
		t.Fatalf("email mismatch: %v", created)
	}

	// Get by id.
	got := mustDo(t, srv, http.MethodGet, "/users/"+id, "", http.StatusOK)
	if got["name"] != "Alice" {
		t.Fatalf("name mismatch: %v", got)
	}

	// List: query.Result envelope {items,total,page,rowsPerPage}.
	listResp := doReq(t, srv, http.MethodGet, "/users", "")
	assertStatus(t, listResp, http.StatusOK)
	var list struct {
		Items []map[string]any `json:"items"`
		Total int              `json:"total"`
	}
	decode(t, listResp, &list)
	if len(list.Items) != 1 || list.Total != 1 {
		t.Fatalf("expected 1 user, got items=%d total=%d", len(list.Items), list.Total)
	}

	// Update (partial: name only) preserves email.
	updated := mustDo(t, srv, http.MethodPut, "/users/"+id, `{"name":"Alicia"}`, http.StatusOK)
	if updated["name"] != "Alicia" {
		t.Fatalf("update did not apply: %v", updated)
	}
	if updated["email"] != "alice@example.com" {
		t.Fatalf("update should preserve email: %v", updated)
	}

	// Validation: a malformed email -> 400 invalid_argument.
	bad := doReq(t, srv, http.MethodPost, "/users", `{"email":"not-an-email","name":"x"}`)
	assertStatus(t, bad, http.StatusBadRequest)
	var errBody map[string]any
	decode(t, bad, &errBody)
	if errBody["code"] != "invalid_argument" {
		t.Fatalf("expected invalid_argument, got %v", errBody)
	}

	// Duplicate email -> 409 already_exists (unique index).
	dup := doReq(t, srv, http.MethodPost, "/users", `{"email":"alice@example.com","name":"Clone"}`)
	assertStatus(t, dup, http.StatusConflict)

	// Delete -> 204, then 404.
	assertStatus(t, doReq(t, srv, http.MethodDelete, "/users/"+id, ""), http.StatusNoContent)
	assertStatus(t, doReq(t, srv, http.MethodGet, "/users/"+id, ""), http.StatusNotFound)
}
