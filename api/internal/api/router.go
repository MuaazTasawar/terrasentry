package api

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/MuaazTasawar/terrasentry/api/internal/notify"
)

// NewRouter builds the full route table for the API server.
func NewRouter(pool *pgxpool.Pool, notifier *notify.Notifier) *gin.Engine {
	router := gin.New()
	router.Use(RequestLogger(), ErrorHandler())

	h := NewHandler(pool, notifier)

	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	v1 := router.Group("/api/v1")
	{
		v1.POST("/scans", h.CreateScan)
		v1.GET("/scans/pending", h.ListPendingScans)
		v1.POST("/scans/:id/decision", h.DecideScan)

		v1.POST("/devices", h.RegisterDevice)

		v1.GET("/drift-events", h.ListDriftEvents)
	}

	return router
}
