package main

import (
    "context"
    "log"
    "net/http"
    "os"
    "os/signal"
    "syscall"
    "time"

    "github.com/joho/godotenv"
)

func main() {
    _ = godotenv.Load("config.env")

    cfg := LoadConfig()
    pool := InitDB(cfg)
    defer pool.Close()

    repo := NewWalletRepository(pool)
    r := SetupRouter(repo)

    srv := &http.Server{
        Addr:    ":" + cfg.ServerPort,
        Handler: r,
    }

    go func() {
       if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            log.Fatalf("listen: %v", err)
        }
    }()
    log.Printf("server listening on :%s", cfg.ServerPort)

    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit
    log.Println("shutdown signal received")

    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("server shutdown: %v", err)
	}

	log.Println("server exited")
}