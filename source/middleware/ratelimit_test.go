package middleware

import (
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func TestRateLimiter_Allow(t *testing.T) {
	limiter := NewRateLimiter(2, time.Minute) // Allow 2 requests per minute
	clientIP := "192.168.1.1"

	// First request should be allowed
	if !limiter.Allow(clientIP) {
		t.Error("First request should be allowed")
	}

	// Second request should be allowed
	if !limiter.Allow(clientIP) {
		t.Error("Second request should be allowed")
	}

	// Third request should be denied
	if limiter.Allow(clientIP) {
		t.Error("Third request should be denied")
	}

	// Different IP should be allowed
	if !limiter.Allow("192.168.1.2") {
		t.Error("Different IP should be allowed")
	}
}

func TestRateLimiter_WindowExpiry(t *testing.T) {
	limiter := NewRateLimiter(1, 100*time.Millisecond) // Very short window for testing
	clientIP := "192.168.1.1"

	// First request should be allowed
	if !limiter.Allow(clientIP) {
		t.Error("First request should be allowed")
	}

	// Second request should be denied
	if limiter.Allow(clientIP) {
		t.Error("Second request should be denied")
	}

	// Wait for window to expire
	time.Sleep(150 * time.Millisecond)

	// Request should be allowed again
	if !limiter.Allow(clientIP) {
		t.Error("Request after window expiry should be allowed")
	}
}

func TestGetClientIP(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name     string
		headers  map[string]string
		expected string
	}{
		{
			name: "X-Forwarded-For header",
			headers: map[string]string{
				"X-Forwarded-For": "192.168.1.1, 10.0.0.1",
			},
			expected: "192.168.1.1",
		},
		{
			name: "X-Real-IP header",
			headers: map[string]string{
				"X-Real-IP": "192.168.1.2",
			},
			expected: "192.168.1.2",
		},
		{
			name:     "No special headers",
			headers:  map[string]string{},
			expected: "", // Will use ClientIP() which depends on the request
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			// Create a request
			req := httptest.NewRequest("GET", "/", nil)
			for k, v := range tt.headers {
				req.Header.Set(k, v)
			}
			c.Request = req

			result := getClientIP(c)

			if tt.expected != "" && result != tt.expected {
				t.Errorf("getClientIP() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestRateLimitMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	limiter := NewRateLimiter(1, time.Minute) // Very restrictive for testing

	r := gin.New()
	r.Use(RateLimitMiddleware(limiter))
	r.GET("/test", func(c *gin.Context) {
		c.String(200, "OK")
	})

	// First request should succeed
	w1 := httptest.NewRecorder()
	req1 := httptest.NewRequest("GET", "/test", nil)
	req1.RemoteAddr = "192.168.1.1:8080"
	r.ServeHTTP(w1, req1)

	if w1.Code != 200 {
		t.Errorf("First request: expected status 200, got %d", w1.Code)
	}

	// Second request from same IP should be rate limited
	w2 := httptest.NewRecorder()
	req2 := httptest.NewRequest("GET", "/test", nil)
	req2.RemoteAddr = "192.168.1.1:8080"
	r.ServeHTTP(w2, req2)

	if w2.Code != 429 {
		t.Errorf("Second request: expected status 429, got %d", w2.Code)
	}

	// Request from different IP should succeed
	w3 := httptest.NewRecorder()
	req3 := httptest.NewRequest("GET", "/test", nil)
	req3.RemoteAddr = "192.168.1.2:8080"
	r.ServeHTTP(w3, req3)

	if w3.Code != 200 {
		t.Errorf("Different IP request: expected status 200, got %d", w3.Code)
	}
}
