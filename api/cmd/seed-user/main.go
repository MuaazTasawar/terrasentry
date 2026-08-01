// Command seed-user creates (or updates the password of) a single user in
// the `users` table, so there's a way to log in before any admin UI exists.
//
// Usage:
//
//    SEED_EMAIL=oncall@example.com SEED_PASSWORD=changeme go run ./cmd/seed-user
package main

import (
    "context"
    "log"
    "os"

    "golang.org/x/crypto/bcrypt"

    "github.com/MuaazTasawar/terrasentry/api/internal/config"
    "github.com/MuaazTasawar/terrasentry/api/internal/db"
)

func main() {
    email := os.Getenv("SEED_EMAIL")
    password := os.Getenv("SEED_PASSWORD")
    if email == "" || password == "" {
        log.Fatal("seed-user: SEED_EMAIL and SEED_PASSWORD environment variables are required")
    }

    cfg := config.Load()

    pool, err := db.NewPool(cfg.DatabaseURL)
    if err != nil {
        log.Fatalf("seed-user: database connection failed: %v", err)
    }
    defer pool.Close()

    hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    if err != nil {
        log.Fatalf("seed-user: failed to hash password: %v", err)
    }

    ctx := context.Background()
    _, err = pool.Exec(ctx, `
        INSERT INTO users (email, password_hash)
        VALUES ($1, $2)
        ON CONFLICT (email) DO UPDATE SET password_hash = EXCLUDED.password_hash
    `, email, string(hash))
    if err != nil {
        log.Fatalf("seed-user: failed to upsert user: %v", err)
    }

    log.Printf("seed-user: user %q is ready to log in", email)
}