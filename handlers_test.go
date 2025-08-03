package main

import (
    "context"
    "bytes"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"

    "github.com/google/uuid"
)

func TestPostDepositAndGetBalance(t *testing.T) {
    repo, cleanup := setupRepository(t)
    defer cleanup()
    router := SetupRouter(repo)

    id := uuid.New().String()
    body := OperationRequest{
        WalletID:      id,
        OperationType: "DEPOSIT",
        Amount:        150,
    }
    payload, _ := json.Marshal(body)
    req, _ := http.NewRequest(http.MethodPost, "/api/v1/wallet", bytes.NewBuffer(payload))
    req.Header.Set("Content-Type", "application/json")
    w := httptest.NewRecorder()
    router.ServeHTTP(w, req)
    if w.Code != http.StatusOK {
        t.Fatalf("expected status 200, got %d", w.Code)
    }
    var resp BalanceResponse
    if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
        t.Fatalf("failed to decode response: %v", err)
    }
    if resp.Balance != 150 {
        t.Fatalf("expected balance 150, got %d", resp.Balance)
    }

    req2, _ := http.NewRequest(http.MethodGet, "/api/v1/wallets/"+id, nil)
    w2 := httptest.NewRecorder()
    router.ServeHTTP(w2, req2)
    if w2.Code != http.StatusOK {
        t.Fatalf("expected status 200, got %d", w2.Code)
    }
    var resp2 BalanceResponse
    if err := json.NewDecoder(w2.Body).Decode(&resp2); err != nil {
        t.Fatalf("failed to decode response: %v", err)
    }
    if resp2.Balance != 150 {
        t.Fatalf("expected balance 150, got %d", resp2.Balance)
    }
}

func TestPostWithdrawInsufficient(t *testing.T) {
    repo, cleanup := setupRepository(t)
    defer cleanup()
    router := SetupRouter(repo)
    id := uuid.New().String()
    if _, err := repo.Deposit(context.Background(), uuid.MustParse(id), 50); err != nil {
        t.Fatalf("setup deposit failed: %v", err)
    }
    
    body := OperationRequest{WalletID: id, OperationType: "WITHDRAW", Amount: 100}
    payload, _ := json.Marshal(body)
    req, _ := http.NewRequest(http.MethodPost, "/api/v1/wallet", bytes.NewBuffer(payload))
    req.Header.Set("Content-Type", "application/json")
    w := httptest.NewRecorder()
    router.ServeHTTP(w, req)
    if w.Code != http.StatusConflict {
        t.Fatalf("expected status 409, got %d", w.Code)
    }
    bal, err := repo.GetBalance(context.Background(), uuid.MustParse(id))
    if err != nil {
        t.Fatalf("get balance failed: %v", err)
    }
    if bal != 50 {
        t.Fatalf("expected balance 50, got %d", bal)
    }
}