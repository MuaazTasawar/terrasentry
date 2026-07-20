package main

import (
	"log"

	"github.com/MuaazTasawar/terrasentry/api/internal/api"
	"github.com/MuaazTasawar/terrasentry/api/internal/config"
	"github.com/MuaazTasawar/terrasentry/api/internal/db"
	"github.com/MuaazTasawar/terrasentry/api/internal/notify"
)

func main() {
	cfg := config.Load()

	pool, err := db.NewPool(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("database connection failed: %v", err)
	}
	defer pool.Close()

	log.Println("connected to database")

	notifier := notify.NewNotifier(cfg.FCMServerKey, pool)
	router := api.NewRouter(pool, notifier, cfg)

	log.Printf("terrasentry api listening on :%s (%s)", cfg.Port, cfg.Env)
	if err := router.Run(":" + cfg.Port); err != nil {
		log.Fatalf("server failed to start: %v", err)
	}
}
