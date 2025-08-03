package main

import (
    "context"
    "errors"
    "fmt"

    "github.com/google/uuid"
    "github.com/jackc/pgx/v5"
    "github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("wallet not found")
var ErrInsufficientFunds = errors.New("insufficient funds")

type WalletRepository struct {
    pool *pgxpool.Pool
}

func NewWalletRepository(pool *pgxpool.Pool) *WalletRepository {
    return &WalletRepository{pool: pool}
}

func (r *WalletRepository) GetBalance(ctx context.Context, id uuid.UUID) (int64, error) {
    var balance int64
    err := r.pool.QueryRow(ctx, `SELECT balance FROM wallets WHERE id = $1`, id).Scan(&balance)
    if err != nil {
        if errors.Is(err, pgx.ErrNoRows) {
            return 0, ErrNotFound
        }
        return 0, err
    }
    return balance, nil
}

func (r *WalletRepository) Deposit(ctx context.Context, id uuid.UUID, amount int64) (int64, error) {
    if amount <= 0 {
        return 0, fmt.Errorf("deposit amount must be positive")
    }
    var newBalance int64
    err := r.pool.QueryRow(ctx, `
        INSERT INTO wallets (id, balance)
        VALUES ($1, $2)
        ON CONFLICT (id) DO UPDATE
        SET balance = wallets.balance + EXCLUDED.balance
        RETURNING balance;
    `, id, amount).Scan(&newBalance)
    if err != nil {
        return 0, err
    }
    return newBalance, nil
}

func (r *WalletRepository) Withdraw(ctx context.Context, id uuid.UUID, amount int64) (int64, error) {
    if amount <= 0 {
        return 0, fmt.Errorf("withdraw amount must be positive")
    }
    var newBalance int64
    err := r.pool.QueryRow(ctx, `
        UPDATE wallets
        SET balance = balance - $2
        WHERE id = $1 AND balance >= $2
        RETURNING balance;
    `, id, amount).Scan(&newBalance)
    if err != nil {
        if errors.Is(err, pgx.ErrNoRows) {
            _, findErr := r.GetBalance(ctx, id)
            if findErr == ErrNotFound {
                return 0, ErrNotFound
            }
            return 0, ErrInsufficientFunds
        }
        return 0, err
    }
    return newBalance, nil
}

func (r *WalletRepository) ApplyOperation(ctx context.Context, id uuid.UUID, operationType string, amount int64) (int64, error) {
    switch operationType {
    case "DEPOSIT":
        return r.Deposit(ctx, id, amount)
    case "WITHDRAW":
        return r.Withdraw(ctx, id, amount)
    default:
        return 0, fmt.Errorf("unknown operation type: %s", operationType)
    }
}