// internal/adapters/outbound/data/postgres/repository/db/connection.go
package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Client struct {
	pool *pgxpool.Pool
}

type Config struct {
	DatabaseURL     string
	MaxConns        int32
	MinConns        int32
	MaxConnLifetime time.Duration
	MaxConnIdleTime time.Duration
}

func NewClient(cfg Config) (*Client, error) {
	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL não pode estar vazia")
	}

	poolConfig, err := pgxpool.ParseConfig(cfg.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("erro ao parsear DATABASE_URL: %w", err)
	}

	// Supabase usa pgbouncer em modo transaction
	poolConfig.ConnConfig.DefaultQueryExecMode = pgx.QueryExecModeSimpleProtocol

	// Configurações de pool
	poolConfig.MaxConns = cfg.MaxConns
	poolConfig.MinConns = cfg.MinConns
	poolConfig.MaxConnLifetime = cfg.MaxConnLifetime
	poolConfig.MaxConnIdleTime = cfg.MaxConnIdleTime

	pool, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
	if err != nil {
		return nil, fmt.Errorf("falha ao criar pool de conexões: %w", err)
	}

	// Testa a conexão
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("falha ao pingar o banco: %w", err)
	}

	return &Client{pool: pool}, nil
}

func (c *Client) Close() {
	if c.pool != nil {
		c.pool.Close()
	}
}

func (c *Client) Pool() *pgxpool.Pool {
	return c.pool
}

// Métodos auxiliares para facilitar o uso
func (c *Client) Ping(ctx context.Context) error {
	return c.pool.Ping(ctx)
}

func (c *Client) BeginTx(ctx context.Context) (pgx.Tx, error) {
	return c.pool.Begin(ctx)
}
