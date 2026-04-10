package handler

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/coder/websocket"

	"github.com/gin-gonic/gin"

	"github.com/chaosplane-hq/chaosplane-platform/apps/api/internal/service"
)

const (
	maxSubscriptionsPerConn = 10
	statusPollInterval      = 2 * time.Second
)

type WebSocketHandler struct {
	svc *service.ExperimentService
}

func NewWebSocketHandler(svc *service.ExperimentService) *WebSocketHandler {
	return &WebSocketHandler{svc: svc}
}

func (h *WebSocketHandler) ExperimentStatus(c *gin.Context) {
	name := c.Param("name")
	namespace := c.DefaultQuery("namespace", "default")

	token := c.Query("token")
	if token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing token query parameter"})
		return
	}

	conn, err := websocket.Accept(c.Writer, c.Request, &websocket.AcceptOptions{
		OriginPatterns: []string{"*"},
	})
	if err != nil {
		slog.Error("websocket accept failed", "error", err)
		return
	}
	defer conn.Close(websocket.StatusNormalClosure, "")

	ctx, cancel := context.WithCancel(c.Request.Context())
	defer cancel()

	h.streamStatus(ctx, conn, namespace, name)
}

func (h *WebSocketHandler) streamStatus(ctx context.Context, conn *websocket.Conn, namespace, name string) {
	ticker := time.NewTicker(statusPollInterval)
	defer ticker.Stop()

	var lastPhase string

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			resp, err := h.svc.Get(ctx, namespace, name)
			if err != nil {
				writeWSError(ctx, conn, err.Error())
				return
			}

			if resp.Phase == lastPhase {
				continue
			}
			lastPhase = resp.Phase

			data, _ := json.Marshal(resp)
			if err := conn.Write(ctx, websocket.MessageText, data); err != nil {
				return
			}

			if resp.Phase == "Completed" || resp.Phase == "Failed" || resp.Phase == "Aborted" {
				return
			}
		}
	}
}

func writeWSError(ctx context.Context, conn *websocket.Conn, msg string) {
	data, _ := json.Marshal(map[string]string{"error": msg})
	_ = conn.Write(ctx, websocket.MessageText, data)
}

// ConnectionLimiter tracks active WebSocket subscriptions per connection.
type ConnectionLimiter struct {
	mu    sync.Mutex
	conns map[string]int
}

func NewConnectionLimiter() *ConnectionLimiter {
	return &ConnectionLimiter{conns: make(map[string]int)}
}

func (cl *ConnectionLimiter) Acquire(key string) bool {
	cl.mu.Lock()
	defer cl.mu.Unlock()
	if cl.conns[key] >= maxSubscriptionsPerConn {
		return false
	}
	cl.conns[key]++
	return true
}

func (cl *ConnectionLimiter) Release(key string) {
	cl.mu.Lock()
	defer cl.mu.Unlock()
	cl.conns[key]--
	if cl.conns[key] <= 0 {
		delete(cl.conns, key)
	}
}
