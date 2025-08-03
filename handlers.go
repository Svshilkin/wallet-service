package main

import (
    "context"
    "net/http"
    "time"

    "github.com/gin-gonic/gin"
    "github.com/google/uuid"
)

type OperationRequest struct {
    WalletID      string `json:"valletId" binding:"required"`
    OperationType string `json:"operationType" binding:"required"`
    Amount        int64  `json:"amount" binding:"required"`
}

type BalanceResponse struct {
    WalletID string `json:"walletId"`
    Balance  int64  `json:"balance"`
}

func SetupRouter(repo *WalletRepository) *gin.Engine {
    router := gin.New()
    router.Use(gin.Logger(), gin.Recovery())

    router.POST("/api/v1/wallet", func(c *gin.Context) {
        var req OperationRequest
        if err := c.ShouldBindJSON(&req); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
            return
        }
        id, err := uuid.Parse(req.WalletID)
        if err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "invalid wallet ID"})
            return
        }
        if req.Amount <= 0 {
            c.JSON(http.StatusBadRequest, gin.H{"error": "amount must be positive"})
            return
        }
        ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
        defer cancel()
        newBalance, err := repo.ApplyOperation(ctx, id, req.OperationType, req.Amount)
        if err != nil {
            switch err {
            case ErrNotFound:
                c.JSON(http.StatusNotFound, gin.H{"error": "wallet not found"})
            case ErrInsufficientFunds:
                c.JSON(http.StatusConflict, gin.H{"error": "insufficient funds"})
            default:
                c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
            }
            return
        }
        c.JSON(http.StatusOK, BalanceResponse{WalletID: id.String(), Balance: newBalance})
    })

    router.GET("/api/v1/wallets/:id", func(c *gin.Context) {
        idStr := c.Param("id")
        id, err := uuid.Parse(idStr)
        if err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "invalid wallet ID"})
            return
        }
        ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
        defer cancel()
        balance, err := repo.GetBalance(ctx, id)
        if err != nil {
            if err == ErrNotFound {
                c.JSON(http.StatusNotFound, gin.H{"error": "wallet not found"})
            } else {
                c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            }
            return
        }
        c.JSON(http.StatusOK, BalanceResponse{WalletID: id.String(), Balance: balance})
    })
    return router
}