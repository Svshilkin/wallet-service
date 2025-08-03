package main

import (
    "fmt"
    "os"
    "strconv"
)

type Config struct {
    DBHost         string
    DBPort         string
    DBUser         string
    DBPassword     string
    DBName         string
    DBSSLMode      string
    ServerPort     string
    DBMaxOpenConns int
    DBMaxIdleConns int
}

func mustGet(key string) string {
    if val, ok := os.LookupEnv(key); ok && val != "" {
        return val
    }
    panic(fmt.Sprintf("env %s not set", key))
}

func mustGetInt(key string) int {
    v := mustGet(key)
    n, err := strconv.Atoi(v)
    if err != nil || n < 0 {
        panic(fmt.Sprintf("env %s must be positive int, got %s", key, v))
    }
    return n
}

func LoadConfig() Config {
    return Config{
        DBHost:         mustGet("DB_HOST"),
        DBPort:         mustGet("DB_PORT"),
        DBUser:         mustGet("DB_USER"),
        DBPassword:     mustGet("DB_PASSWORD"),
        DBName:         mustGet("DB_NAME"),
        DBSSLMode:      mustGet("DB_SSLMODE"),
        ServerPort:     mustGet("SERVER_PORT"),
        DBMaxOpenConns: mustGetInt("DB_MAX_OPEN_CONNS"),
        DBMaxIdleConns: mustGetInt("DB_MAX_IDLE_CONNS"),
    }
}
