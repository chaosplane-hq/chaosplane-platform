package database

import "embed"

// MigrationsFS holds the embedded SQL migration files applied by the migrate
// command. Files follow the golang-migrate naming convention
// (NNNNNN_name.up.sql / NNNNNN_name.down.sql).
//
//go:embed migrations/*.sql
var MigrationsFS embed.FS
