package migrations

import (
	"context"

	"github.com/nicolasbonnici/gorest/database"
	"github.com/nicolasbonnici/gorest/migrations"
)

func GetMigrations() migrations.MigrationSource {
	builder := migrations.NewMigrationBuilder("gorest-metrics")

	builder.Add(
		"20260207000003000",
		"create_metrics_table",
		func(ctx context.Context, db database.Database) error {
			if err := migrations.SQL(ctx, db, migrations.DialectSQL{
				Postgres: `CREATE TABLE IF NOT EXISTS metrics (
					id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
					resource TEXT NOT NULL,
					resource_id UUID NOT NULL,
					key VARCHAR(255) NOT NULL,
					value INTEGER NOT NULL DEFAULT 0,
					created_at TIMESTAMP(0) WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
					CONSTRAINT unique_resource_metric UNIQUE (resource, resource_id, key)
				)`,
				MySQL: `CREATE TABLE IF NOT EXISTS metrics (
					id CHAR(36) PRIMARY KEY,
					resource VARCHAR(255) NOT NULL,
					resource_id CHAR(36) NOT NULL,
					` + "`key`" + ` VARCHAR(255) NOT NULL,
					value INT NOT NULL DEFAULT 0,
					created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
					UNIQUE KEY unique_resource_metric (resource, resource_id, ` + "`key`" + `),
					INDEX idx_metrics_resource (resource, resource_id, ` + "`key`" + `),
					INDEX idx_metrics_key (` + "`key`" + `, created_at),
					INDEX idx_metrics_resource_id (resource_id)
				) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,
				SQLite: `CREATE TABLE IF NOT EXISTS metrics (
					id TEXT PRIMARY KEY,
					resource TEXT NOT NULL,
					resource_id TEXT NOT NULL,
					key TEXT NOT NULL,
					value INTEGER NOT NULL DEFAULT 0,
					created_at TEXT NOT NULL DEFAULT (datetime('now')),
					UNIQUE (resource, resource_id, key)
				)`,
			}); err != nil {
				return err
			}

			// Create indexes for Postgres and SQLite
			if db.DriverName() == "postgres" {
				if err := migrations.SQL(ctx, db, migrations.DialectSQL{
					Postgres: `CREATE INDEX IF NOT EXISTS idx_metrics_resource ON metrics(resource, resource_id, key)`,
				}); err != nil {
					return err
				}
				if err := migrations.SQL(ctx, db, migrations.DialectSQL{
					Postgres: `CREATE INDEX IF NOT EXISTS idx_metrics_key ON metrics(key, created_at)`,
				}); err != nil {
					return err
				}
				if err := migrations.CreateIndex(ctx, db, "idx_metrics_resource_id", "metrics", "resource_id"); err != nil {
					return err
				}
			}

			if db.DriverName() == "sqlite" {
				if err := migrations.SQL(ctx, db, migrations.DialectSQL{
					SQLite: `CREATE INDEX IF NOT EXISTS idx_metrics_resource ON metrics(resource, resource_id, key)`,
				}); err != nil {
					return err
				}
				if err := migrations.SQL(ctx, db, migrations.DialectSQL{
					SQLite: `CREATE INDEX IF NOT EXISTS idx_metrics_key ON metrics(key, created_at)`,
				}); err != nil {
					return err
				}
				if err := migrations.CreateIndex(ctx, db, "idx_metrics_resource_id", "metrics", "resource_id"); err != nil {
					return err
				}
			}

			return nil
		},
		func(ctx context.Context, db database.Database) error {
			if db.DriverName() == "postgres" || db.DriverName() == "sqlite" {
				_ = migrations.DropIndex(ctx, db, "idx_metrics_resource", "metrics")
				_ = migrations.DropIndex(ctx, db, "idx_metrics_key", "metrics")
				_ = migrations.DropIndex(ctx, db, "idx_metrics_resource_id", "metrics")
			}

			return migrations.DropTableIfExists(ctx, db, "metrics")
		},
	)

	return builder.Build()
}
