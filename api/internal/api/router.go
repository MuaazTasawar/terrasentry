package api

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/MuaazTasawar/terrasentry/api/internal/config"
	"github.com/MuaazTasawar/terrasentry/api/internal/notify"
)

// NewRouter builds the full route table for the API server.
func NewRouter(pool *pgxpool.Pool, notifier *notify.Notifier, cfg *config.Config) *gin.Engine {
	router := gin.New()
	router.Use(RequestLogger(), ErrorHandler(), CORS())

	h := NewHandler(pool, notifier, cfg.JWTSecret, cfg.JWTExpiryHours)
	authRequired := AuthRequired(cfg.JWTSecret)

	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	v1 := router.Group("/api/v1")
	{
		v1.POST("/auth/login", h.Login)

		v1.POST("/scans", h.CreateScan)
		v1.GET("/scans/pending", h.ListPendingScans)
		v1.GET("/scans", authRequired, h.ListAllScans)
		v1.POST("/scans/:id/decision", authRequired, h.DecideScan)

		v1.POST("/devices", authRequired, h.RegisterDevice)

		v1.GET("/drift-events", h.ListDriftEvents)

		v1.GET("/audit", authRequired, h.ListApprovalAudit)
	}

	return router
}
