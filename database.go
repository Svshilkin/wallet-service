package main

import (
    "context"
    "fmt"
    "log"
    "time"

    "github.com/jackc/pgx/v5/pgxpool"
)

func InitDB(cfg Config) *pgxpool.Pool {
    connString := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s", cfg.DBUser, cfg.DBPassword, cfg.DBHost, cfg.DBPort, cfg.DBName, cfg.DBSSLMode)
    poolConfig, err := pgxpool.ParseConfig(connString)
    if err != nil {
        log.Fatalf("failed to parse database config: %v", err)
    }
    
    poolConfig.MaxConns = int32(cfg.DBMaxOpenConns)
    poolConfig.MaxConnIdleTime = time.Minute * 5
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()
    pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
    if err != nil {
        log.Fatalf("unable to connect to database: %v", err)
    }

    if err := createSchema(ctx, pool); err != nil {
        pool.Close()
        log.Fatalf("failed to initialize database schema: %v", err)
    }
    return pool
}

func createSchema(ctx context.Context, pool *pgxpool.Pool) error {
    schema := `
CREATE TABLE IF NOT EXISTS wallets (
  id UUID PRIMARY KEY,
  balance BIGINT NOT NULL DEFAULT 0
);
`
    _, err := pool.Exec(ctx, schema)
    return err
}