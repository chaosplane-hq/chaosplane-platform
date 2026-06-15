package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestWebSocketHandler_MissingToken(t *testing.T) {
	h := NewWebSocketHandler(nil, nil, nil)

	w := httptest.NewRecorder()
	_, r := gin.CreateTestContext(w)
	r.GET("/ws/experiments/:name", h.ExperimentStatus)
	req := httptest.NewRequest(http.MethodGet, "/ws/experiments/test-exp", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestConnectionLimiter(t *testing.T) {
	cl := NewConnectionLimiter()

	for i := 0; i < maxSubscriptionsPerConn; i++ {
		if !cl.Acquire("key1") {
			t.Fatalf("expected acquire to succeed at %d", i)
		}
	}

	if cl.Acquire("key1") {
		t.Error("expected acquire to fail at max")
	}

	cl.Release("key1")
	if !cl.Acquire("key1") {
		t.Error("expected acquire to succeed after release")
	}

	if !cl.Acquire("key2") {
		t.Error("expected acquire for different key to succeed")
	}
}
