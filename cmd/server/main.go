package main

import (
    "log"
    "os"

    "github.com/joho/godotenv"

    "online-disk-server/internal/server"
)

func main() {
    // Load .env if present (optional for local dev)
    _ = godotenv.Load()

    // Start HTTP server
    if err := server.Run(); err != nil {
        log.Printf("server exited with error: %v", err)
        os.Exit(1)
    }
}
