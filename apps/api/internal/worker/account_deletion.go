package worker

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/chaosplane-hq/chaosplane-platform/apps/api/internal/database"
	"github.com/chaosplane-hq/chaosplane-platform/apps/api/internal/service"
)

type AccountDeletionWorker struct {
	svc      *service.AccountDeletionService
	interval time.Duration
}

func NewAccountDeletionWorker(pool *database.Pool) *AccountDeletionWorker {
	return &AccountDeletionWorker{
		svc:      service.NewAccountDeletionService(pool),
		interval: 1 * time.Hour,
	}
}

func (w *AccountDeletionWorker) Start(ctx context.Context) {
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	w.run(ctx)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			w.run(ctx)
		}
	}
}

func (w *AccountDeletionWorker) run(ctx context.Context) {
	count, err := w.svc.ExecuteExpired(ctx)
	if err != nil {
		slog.Error("account deletion worker failed", "error", err)
		return
	}
	if count > 0 {
		slog.Info(fmt.Sprintf("executed %d expired account deletions", count))
	}
}
