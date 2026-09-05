package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func TestRateLimiterRejectsRequestsAfterLimit(t *testing.T) {
	limiter := &rateLimiter{window: time.Minute, limit: 2, buckets: map[string]rateLimitBucket{}}
	if !limiter.allow("127.0.0.1") || !limiter.allow("127.0.0.1") {
		t.Fatal("the first two requests should be allowed")
	}
	if limiter.allow("127.0.0.1") {
		t.Fatal("the request after the limit should be rejected")
	}
	if !limiter.allow("192.0.2.1") {
		t.Fatal("a different client should have its own bucket")
	}
}

func TestMaxBodyBytesReturnsRequestEntityTooLarge(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/upload", MaxBodyBytes(4), func(c *gin.Context) {
		body := make([]byte, 8)
		_, err := c.Request.Body.Read(body)
		if err == nil {
			c.Status(http.StatusOK)
			return
		}
		c.Status(http.StatusRequestEntityTooLarge)
	})

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/upload", strings.NewReader("12345678"))
	router.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusRequestEntityTooLarge)
	}
}
