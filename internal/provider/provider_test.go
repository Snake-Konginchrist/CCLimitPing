package provider

import (
	"context"
	"io"
	"net/http"
	"strings"
	"syscall"
	"testing"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestDoGetRetriesTransientNetworkError(t *testing.T) {
	oldClient := http.DefaultClient
	defer func() { http.DefaultClient = oldClient }()

	attempts := 0
	http.DefaultClient = &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		attempts++
		if attempts == 1 {
			return nil, syscall.ECONNRESET
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(`{"ok":true}`)),
			Request:    req,
		}, nil
	})}

	body, status, err := doGet(context.Background(), "token", func(token string) (*http.Request, error) {
		return http.NewRequest(http.MethodGet, "https://example.test/usage", nil)
	})
	if err != nil {
		t.Fatalf("doGet: %v", err)
	}
	if status != http.StatusOK {
		t.Fatalf("status = %d, want %d", status, http.StatusOK)
	}
	if string(body) != `{"ok":true}` {
		t.Fatalf("body = %q", body)
	}
	if attempts != 2 {
		t.Fatalf("attempts = %d, want 2", attempts)
	}
}

func TestDoGetDoesNotRetryUnauthorized(t *testing.T) {
	oldClient := http.DefaultClient
	defer func() { http.DefaultClient = oldClient }()

	attempts := 0
	http.DefaultClient = &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		attempts++
		return &http.Response{
			StatusCode: http.StatusUnauthorized,
			Body:       io.NopCloser(strings.NewReader(`{"error":"expired"}`)),
			Request:    req,
		}, nil
	})}

	body, status, err := doGet(context.Background(), "token", func(token string) (*http.Request, error) {
		return http.NewRequest(http.MethodGet, "https://example.test/usage", nil)
	})
	if err != nil {
		t.Fatalf("doGet: %v", err)
	}
	if status != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", status, http.StatusUnauthorized)
	}
	if string(body) != `{"error":"expired"}` {
		t.Fatalf("body = %q", body)
	}
	if attempts != 1 {
		t.Fatalf("attempts = %d, want 1", attempts)
	}
}
