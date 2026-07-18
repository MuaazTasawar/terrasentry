package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/MuaazTasawar/terrasentry/api/internal/config"
	"github.com/MuaazTasawar/terrasentry/api/internal/db"
)

func main() {
	cfg := config.Load()

	pool, err := db.NewPool(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("database connection failed: %v", err)
	}
	defer pool.Close()

	log.Println("connected to database")

	router := gin.Default()

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
			"env":    cfg.Env,
		})
	})

	// Route groups (handlers.go) and middleware (middleware.go) are wired in Phase 4.

	log.Printf("terrasentry api listening on :%s (%s)", cfg.Port, cfg.Env)
	if err := router.Run(":" + cfg.Port); err != nil {
		log.Fatalf("server failed to start: %v", err)
	}
}
