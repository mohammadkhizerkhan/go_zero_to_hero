package test_context

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestHTTPRequestWithContext(t *testing.T) {
	// 1. Create a slow test HTTP server (delays response by 3 seconds)
	slowServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(3 * time.Second)
		w.WriteHeader(http.StatusOK)
	}))
	defer slowServer.Close()

	// 2. Create a context with a strict 1-second timeout
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	// 3. Bind the context to the HTTP Request
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, slowServer.URL, nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	// 4. Execute the HTTP Request
	client := &http.Client{}
	_, err = client.Do(req)

	if err == nil {
		t.Fatal("Expected HTTP request to fail due to timeout, but it succeeded.")
	}
	t.Logf("HTTP Request canceled successfully as expected: %v", err)
}

