package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.Header.Get("Authorization") != "Bearer test-key" {
			t.Errorf("missing or wrong auth header")
		}
		if r.URL.Path != "/api/v1/agent/accounts" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("type") != "cash" {
			t.Errorf("expected type=cash query param")
		}
		w.WriteHeader(200)
		w.Write([]byte(`{"accounts":[]}`))
	}))
	defer server.Close()

	client := New("test-key", server.URL)
	data, err := client.Get("/agent/accounts", map[string]string{"type": "cash"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var resp struct {
		Accounts []any `json:"accounts"`
	}
	json.Unmarshal(data, &resp)
	if len(resp.Accounts) != 0 {
		t.Errorf("expected empty accounts")
	}
}

func TestPostSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("expected JSON content type")
		}
		w.WriteHeader(201)
		w.Write([]byte(`{"created":1,"updated":0}`))
	}))
	defer server.Close()

	client := New("test-key", server.URL)
	body := map[string]any{"accounts": []any{}}
	data, err := client.Post("/agent/accounts/bulk-upsert", body)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var resp struct {
		Created int `json:"created"`
	}
	json.Unmarshal(data, &resp)
	if resp.Created != 1 {
		t.Errorf("expected created=1, got %d", resp.Created)
	}
}

func TestErrorContract(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(403)
		w.Write([]byte(`{"error":"forbidden","message":"Missing required scope: accounts:write","status":403}`))
	}))
	defer server.Close()

	client := New("test-key", server.URL)
	_, err := client.Get("/agent/accounts", nil)
	if err == nil {
		t.Fatal("expected error")
	}
	if err.Error() != "Missing required scope: accounts:write" {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestError401(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(401)
		w.Write([]byte(`{}`))
	}))
	defer server.Close()

	client := New("bad-key", server.URL)
	_, err := client.Get("/agent/accounts", nil)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestEmptyParams(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.RawQuery != "" {
			t.Errorf("expected no query params, got %s", r.URL.RawQuery)
		}
		w.WriteHeader(200)
		w.Write([]byte(`{}`))
	}))
	defer server.Close()

	client := New("test-key", server.URL)
	// Empty string values should not be included.
	client.Get("/agent/accounts", map[string]string{"type": ""})
}
