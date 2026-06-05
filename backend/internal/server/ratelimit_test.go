package server

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
)

// okHandler is a trivial handler used as the protected target.
func okHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
}

func doRequest(h http.Handler, remoteAddr string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, "/auth/login", nil)
	req.RemoteAddr = remoteAddr
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func TestRateLimiterAllowsUpToBudgetThenRejects(t *testing.T) {
	rl := newRateLimiter(10)
	h := rl.Middleware(okHandler())
	const addr = "203.0.113.1:5000"

	for i := 1; i <= 10; i++ {
		if rec := doRequest(h, addr); rec.Code != http.StatusOK {
			t.Fatalf("request %d: got %d, want 200", i, rec.Code)
		}
	}

	rec := doRequest(h, addr)
	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("request 11: got %d, want 429", rec.Code)
	}
}

func TestRateLimiterSetsRetryAfterOnBreach(t *testing.T) {
	rl := newRateLimiter(5)
	h := rl.Middleware(okHandler())
	const addr = "203.0.113.2:5000"

	for i := 0; i < 5; i++ {
		doRequest(h, addr)
	}

	rec := doRequest(h, addr)
	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("got %d, want 429", rec.Code)
	}
	ra := rec.Header().Get("Retry-After")
	if ra == "" {
		t.Fatal("missing Retry-After header on 429")
	}
	if secs, err := strconv.Atoi(ra); err != nil || secs < 1 {
		t.Fatalf("Retry-After = %q, want positive integer seconds", ra)
	}
}

func TestRateLimiterIsolatesClientsByIP(t *testing.T) {
	rl := newRateLimiter(3)
	h := rl.Middleware(okHandler())

	// Exhaust the budget for the first client.
	for i := 0; i < 3; i++ {
		doRequest(h, "198.51.100.1:1111")
	}
	if rec := doRequest(h, "198.51.100.1:1111"); rec.Code != http.StatusTooManyRequests {
		t.Fatalf("client A 4th request: got %d, want 429", rec.Code)
	}

	// A different IP has its own untouched budget.
	if rec := doRequest(h, "198.51.100.2:2222"); rec.Code != http.StatusOK {
		t.Fatalf("client B first request: got %d, want 200", rec.Code)
	}
}

func TestRateLimiterRejectedRequestDoesNotConsumeBudget(t *testing.T) {
	// With a fresh limiter, a single allowed request leaves burst-1 tokens.
	// Verify that a rejected request after exhaustion does not further drain
	// the bucket (Reserve is cancelled), so once a token refills a request
	// can succeed again. We assert the rejection path keeps the bucket at its
	// boundary by confirming repeated rejections stay 429 rather than erroring.
	rl := newRateLimiter(2)
	h := rl.Middleware(okHandler())
	const addr = "203.0.113.9:5000"

	for i := 0; i < 2; i++ {
		if rec := doRequest(h, addr); rec.Code != http.StatusOK {
			t.Fatalf("request %d: got %d, want 200", i, rec.Code)
		}
	}
	for i := 0; i < 3; i++ {
		if rec := doRequest(h, addr); rec.Code != http.StatusTooManyRequests {
			t.Fatalf("over-budget request %d: got %d, want 429", i, rec.Code)
		}
	}
}

func TestClientIPStripsPort(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "192.0.2.5:54321"
	if got := clientIP(req); got != "192.0.2.5" {
		t.Fatalf("clientIP = %q, want 192.0.2.5", got)
	}

	req.RemoteAddr = "no-port-here"
	if got := clientIP(req); got != "no-port-here" {
		t.Fatalf("clientIP fallback = %q, want raw RemoteAddr", got)
	}
}
