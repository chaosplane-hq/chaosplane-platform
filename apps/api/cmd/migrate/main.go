// Command migrate applies database schema migrations embedded in the binary.
//
// It reads DATABASE_URL from the environment and runs all pending "up"
// migrations from the embedded migrations. It is intended to run as a
// Kubernetes Job (ArgoCD sync hook) before the API server rolls out.
package main

import (
	"errors"
	"log/slog"
	"os"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	"github.com/golang-migrate/migrate/v4/source/iofs"

	"github.com/chaosplane-hq/chaosplane-platform/apps/api/internal/database"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		logger.Error("DATABASE_URL is required")
		os.Exit(1)
	}

	source, err := iofs.New(database.MigrationsFS, "migrations")
	if err != nil {
		logger.Error("failed to load embedded migrations", "error", err)
		os.Exit(1)
	}

	m, err := migrate.NewWithSourceInstance("iofs", source, toPgxURL(databaseURL))
	if err != nil {
		logger.Error("failed to initialize migrator", "error", err)
		os.Exit(1)
	}
	defer m.Close()

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		logger.Error("migration failed", "error", err)
		os.Exit(1)
	}

	version, dirty, verr := m.Version()
	if verr != nil && !errors.Is(verr, migrate.ErrNilVersion) {
		logger.Error("failed to read schema version", "error", verr)
		os.Exit(1)
	}

	logger.Info("migrations applied", "version", version, "dirty", dirty)
}

// toPgxURL rewrites a postgres://|postgresql:// connection string to the
// pgx5:// scheme that the golang-migrate pgx/v5 driver expects.
func toPgxURL(url string) string {
	for _, prefix := range []string{"postgresql://", "postgres://"} {
		if rest, ok := strings.CutPrefix(url, prefix); ok {
			return "pgx5://" + rest
		}
	}
	return url
}
