// Package db manages the PostgreSQL connection pool and schema migrations.
package db

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// InitPool creates and validates a pgxpool connection pool from connStr.
// If connStr is empty the function returns nil without error; callers must
// guard against a nil pool before issuing queries.
func InitPool(ctx context.Context, connStr string) *pgxpool.Pool {
	if connStr == "" {
		log.Println("db: AIVEN_DB_URL is not set — running without database")
		return nil
	}

	cfg, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		log.Fatalf("db: parsing connection string: %v", err)
	}

	// Aiven free tier hard limit: 20 concurrent connections.
	// Reserve 5 for GUI tools (DBeaver / pgAdmin) used during development.
	cfg.MaxConns = 15
	cfg.MinConns = 2
	cfg.MaxConnLifetime = 30 * time.Minute
	cfg.MaxConnIdleTime = 10 * time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		log.Fatalf("db: creating pool: %v", err)
	}

	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("db: pinging database: %v", err)
	}

	log.Println("db: connected to PostgreSQL")
	return pool
}

// RunMigrations applies all pending SQL migrations from the db/migrations
// directory in lexicographic order.  Each migration file is executed inside
// its own transaction; a failure rolls back that migration and halts startup.
func RunMigrations(ctx context.Context, pool *pgxpool.Pool) {
	if pool == nil {
		log.Println("db: skipping migrations — no database connection")
		return
	}

	pattern := filepath.Join("db", "migrations", "*.sql")
	files, err := filepath.Glob(pattern)
	if err != nil {
		log.Fatalf("db: globbing migrations: %v", err)
	}
	sort.Strings(files) // ensure 001 < 002 < … order

	for _, f := range files {
		sql, err := os.ReadFile(f)
		if err != nil {
			log.Fatalf("db: reading migration %s: %v", f, err)
		}

		// Skip blank files.
		if strings.TrimSpace(string(sql)) == "" {
			continue
		}

		tx, err := pool.Begin(ctx)
		if err != nil {
			log.Fatalf("db: beginning transaction for %s: %v", f, err)
		}

		if _, err := tx.Exec(ctx, string(sql)); err != nil {
			_ = tx.Rollback(ctx)
			log.Fatalf("db: executing migration %s: %v", f, err)
		}

		if err := tx.Commit(ctx); err != nil {
			log.Fatalf("db: committing migration %s: %v", f, err)
		}

		fmt.Printf("db: applied migration %s\n", filepath.Base(f))
	}
}
