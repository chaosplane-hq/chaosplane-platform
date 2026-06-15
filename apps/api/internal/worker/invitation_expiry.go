package worker

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/chaosplane-hq/chaosplane-platform/apps/api/internal/database"
)

type InvitationExpiryWorker struct {
	pool     *database.Pool
	interval time.Duration
}

func NewInvitationExpiryWorker(pool *database.Pool) *InvitationExpiryWorker {
	return &InvitationExpiryWorker{pool: pool, interval: 5 * time.Minute}
}

func (w *InvitationExpiryWorker) Start(ctx context.Context) {
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	w.expire(ctx)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			w.expire(ctx)
		}
	}
}

func (w *InvitationExpiryWorker) expire(ctx context.Context) {
	cmd, err := w.pool.Conn(ctx).Exec(ctx, `
		UPDATE invitations
		SET status = 'expired'
		WHERE status = 'pending' AND expires_at < now()
	`)
	if err != nil {
		slog.Error("invitation expiry worker failed", "error", err)
		return
	}
	if affected := cmd.RowsAffected(); affected > 0 {
		slog.Info(fmt.Sprintf("expired %d pending invitations", affected))
	}
}
