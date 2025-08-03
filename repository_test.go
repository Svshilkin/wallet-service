package main

import (
    "context"
    "os"
    "sync"
    "testing"
    "time"

    "github.com/google/uuid"
    "github.com/joho/godotenv"
)

func testConfig() Config {
    return LoadConfig()
}

func TestMain(m *testing.M) {
    _ = godotenv.Load("config.env")
    os.Exit(m.Run())
}

func setupRepository(t *testing.T) (*WalletRepository, func()) {
    t.Helper()
    cfg := testConfig()
    pool := InitDB(cfg)
    repo := NewWalletRepository(pool)
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    _, err := pool.Exec(ctx, "TRUNCATE TABLE wallets")
    cancel()
    if err != nil {
        t.Fatalf("failed to truncate wallets table: %v", err)
    }
    cleanup := func() {
        pool.Close()
    }
    return repo, cleanup
}

func TestDepositAndGet(t *testing.T) {
    repo, cleanup := setupRepository(t)
    defer cleanup()
    id := uuid.New()
    ctx := context.Background()
    bal, err := repo.Deposit(ctx, id, 100)
    if err != nil {
        t.Fatalf("deposit failed: %v", err)
    }
    if bal != 100 {
        t.Fatalf("expected balance 100, got %d", bal)
    }
    got, err := repo.GetBalance(ctx, id)
    if err != nil {
        t.Fatalf("get balance failed: %v", err)
    }
    if got != 100 {
        t.Fatalf("expected balance 100, got %d", got)
    }
}

func TestWithdrawInsufficient(t *testing.T) {
    repo, cleanup := setupRepository(t)
    defer cleanup()
    id := uuid.New()
    ctx := context.Background()
    if _, err := repo.Deposit(ctx, id, 50); err != nil {
        t.Fatalf("deposit failed: %v", err)
    }
    if _, err := repo.Withdraw(ctx, id, 100); err != ErrInsufficientFunds {
        t.Fatalf("expected ErrInsufficientFunds, got %v", err)
    }
    bal, err := repo.GetBalance(ctx, id)
    if err != nil {
        t.Fatalf("get balance failed: %v", err)
    }
    if bal != 50 {
        t.Fatalf("expected balance 50, got %d", bal)
    }
}

func TestConcurrentDeposits(t *testing.T) {
    repo, cleanup := setupRepository(t)
    defer cleanup()
    id := uuid.New()
    const numGoroutines = 100
    const depositAmount = int64(10)
    ctx := context.Background()
    var wg sync.WaitGroup
    wg.Add(numGoroutines)
    for i := 0; i < numGoroutines; i++ {
        go func() {
            defer wg.Done()
            if _, err := repo.Deposit(ctx, id, depositAmount); err != nil {
                t.Errorf("deposit error: %v", err)
            }
        }()
    }
    wg.Wait()
    bal, err := repo.GetBalance(ctx, id)
    if err != nil {
        t.Fatalf("get balance failed: %v", err)
    }
    expected := int64(numGoroutines) * depositAmount
    if bal != expected {
        t.Fatalf("expected balance %d, got %d", expected, bal)
    }
}

func TestConcurrentWithdrawals(t *testing.T) {
    repo, cleanup := setupRepository(t)
    defer cleanup()
    id := uuid.New()
    ctx := context.Background()
    initial := int64(200)
    if _, err := repo.Deposit(ctx, id, initial); err != nil {
        t.Fatalf("deposit failed: %v", err)
    }
    const workers = 50
    const perWithdraw = int64(3)
    expectedSuccesses := int(initial / perWithdraw)
    var wg sync.WaitGroup
    wg.Add(workers)
    for i := 0; i < workers; i++ {
        go func() {
            defer wg.Done()
            for {
                _, err := repo.Withdraw(ctx, id, perWithdraw)
                if err == nil {
                    continue
                }
                if err == ErrInsufficientFunds || err == ErrNotFound {
                    return
                }
                t.Errorf("withdraw error: %v", err)
                return
            }
        }()
    }
    wg.Wait()
    bal, err := repo.GetBalance(ctx, id)
    if err != nil {
        t.Fatalf("get balance failed: %v", err)
    }
    expectedBalance := initial - int64(expectedSuccesses)*perWithdraw
    if bal != expectedBalance {
        t.Fatalf("expected balance %d, got %d", expectedBalance, bal)
    }
}