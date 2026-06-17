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

	"github.com/chaosplane-hq/chaosplane-platform/apps/api/internal/database"
	"github.com/chaosplane-hq/chaosplane-platform/apps/api/internal/service"
)

const (
	maxSubscriptionsPerConn = 10
	statusPollInterval      = 2 * time.Second
)

type WebSocketHandler struct {
	svc     *service.ExperimentService
	auth    *service.AuthService
	pool    *database.Pool
	limiter *ConnectionLimiter
}

func NewWebSocketHandler(svc *service.ExperimentService, auth *service.AuthService, pool *database.Pool) *WebSocketHandler {
	return &WebSocketHandler{svc: svc, auth: auth, pool: pool, limiter: NewConnectionLimiter()}
}

func (h *WebSocketHandler) ExperimentStatus(c *gin.Context) {
	id := c.Param("name")

	token := c.Query("token")
	if token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing token query parameter"})
		return
	}
	claims, err := h.auth.ParseAccessToken(token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid access token"})
		return
	}

	if !h.limiter.Acquire(claims.Subject) {
		c.JSON(http.StatusTooManyRequests, gin.H{"error": "too many WebSocket connections"})
		return
	}
	defer h.limiter.Release(claims.Subject)

	actor := service.ActorContext{UserID: claims.Subject, TenantID: claims.TenantID}

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

	h.streamStatus(ctx, conn, actor, id)
}

func (h *WebSocketHandler) streamStatus(ctx context.Context, conn *websocket.Conn, actor service.ActorContext, id string) {
	ticker := time.NewTicker(statusPollInterval)
	defer ticker.Stop()

	var lastPhase string

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			resp, err := h.getScoped(ctx, actor, id)
			if err != nil {
				writeWSError(ctx, conn, err.Error())
				return
			}

			if resp.Status.Phase == lastPhase {
				continue
			}
			lastPhase = resp.Status.Phase

			data, _ := json.Marshal(resp)
			if err := conn.Write(ctx, websocket.MessageText, data); err != nil {
				return
			}

			if resp.Status.Phase == "Completed" || resp.Status.Phase == "Failed" || resp.Status.Phase == "Aborted" {
				return
			}
		}
	}
}

// getScoped opens a request-scoped RLS transaction since the WebSocket route
// bypasses TenantContext middleware, then reads the experiment within it.
func (h *WebSocketHandler) getScoped(ctx context.Context, actor service.ActorContext, id string) (*service.ExperimentResponse, error) {
	tx, err := h.pool.App.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback(context.Background()) }()

	if _, err := tx.Exec(ctx, "SELECT set_config('app.current_tenant_id', $1, true)", actor.TenantID); err != nil {
		return nil, err
	}
	return h.svc.Get(database.WithTx(ctx, tx), actor, id)
}

func writeWSError(ctx context.Context, conn *websocket.Conn, msg string) {
	data, _ := json.Marshal(map[string]string{"error": msg})
	_ = conn.Write(ctx, websocket.MessageText, data)
}

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
