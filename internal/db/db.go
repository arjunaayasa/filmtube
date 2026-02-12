package db

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

type DB struct {
	*sqlx.DB
}

// Connect creates a new database connection
func Connect(databaseURL string) (*DB, error) {
	db, err := sqlx.Connect("postgres", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(0) // Connections live forever

	return &DB{DB: db}, nil
}

// Ping checks if database is reachable
func (d *DB) Ping(ctx context.Context) error {
	return d.DB.PingContext(ctx)
}

// BeginTx starts a transaction
func (d *DB) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sqlx.Tx, error) {
	return d.DB.BeginTxx(ctx, opts)
}

// Close closes the database connection
func (d *DB) Close() error {
	return d.DB.Close()
}
