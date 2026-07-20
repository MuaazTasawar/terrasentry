package api

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"

	"github.com/MuaazTasawar/terrasentry/api/internal/auth"
	"github.com/MuaazTasawar/terrasentry/api/internal/notify"
)

type Handler struct {
	Pool           *pgxpool.Pool
	Notifier       *notify.Notifier
	JWTSecret      string
	JWTExpiryHours int
}

func NewHandler(pool *pgxpool.Pool, notifier *notify.Notifier, jwtSecret string, jwtExpiryHours int) *Handler {
	return &Handler{Pool: pool, Notifier: notifier, JWTSecret: jwtSecret, JWTExpiryHours: jwtExpiryHours}
}

// --- Auth ---

type loginInput struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// Login verifies email/password against the users table and, on success,
// issues a signed JWT for use as a Bearer token on protected endpoints.
func (h *Handler) Login(c *gin.Context) {
	var input loginInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var (
		userID       string
		passwordHash string
	)
	err := h.Pool.QueryRow(c.Request.Context(),
		`SELECT id, password_hash FROM users WHERE email = $1`,
		input.Email,
	).Scan(&userID, &passwordHash)
	if err != nil {
		// Same response whether the user doesn't exist or the password is
		// wrong, so login can't be used to enumerate registered emails.
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid email or password"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(input.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid email or password"})
		return
	}

	token, err := auth.GenerateToken(h.JWTSecret, h.JWTExpiryHours, userID, input.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to issue token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token, "expires_in_hours": h.JWTExpiryHours})
}

// --- Plan Scans ---

type createScanInput struct {
	RepoName    string `json:"repo_name" binding:"required"`
	PlanSummary string `json:"plan_summary" binding:"required"`
	RiskScore   int    `json:"risk_score" binding:"required"`
	RiskLevel   string `json:"risk_level" binding:"required"`
	Reasoning   string `json:"reasoning"`
}

// CreateScan stores a scored plan (called by the risk-scoring service or
// a CI pipeline after it gets a score back) and, if risk is medium/high,
// pushes a mobile notification for approval.
func (h *Handler) CreateScan(c *gin.Context) {
	var input createScanInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	status := "scored"

	var id string
	err := h.Pool.QueryRow(c.Request.Context(),
		`INSERT INTO plan_scans (repo_name, plan_summary, risk_score, risk_level, reasoning, status)
		 VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`,
		input.RepoName, input.PlanSummary, input.RiskScore, input.RiskLevel, input.Reasoning, status,
	).Scan(&id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save scan: " + err.Error()})
		return
	}

	if input.RiskLevel == "medium" || input.RiskLevel == "high" {
		// Non-fatal: notification failure shouldn't fail the scan creation.
		_ = h.Notifier.NotifyRiskyPlan(c.Request.Context(), id, input.RepoName, input.RiskLevel)
	}

	c.JSON(http.StatusCreated, gin.H{"id": id, "status": status})
}

// ListPendingScans returns scans awaiting a decision — this is what the
// Flutter app's home screen polls/fetches.
func (h *Handler) ListPendingScans(c *gin.Context) {
	rows, err := h.Pool.Query(c.Request.Context(), `
		SELECT id, repo_name, plan_summary, risk_score, risk_level, reasoning, status, created_at
		FROM plan_scans
		WHERE status = 'scored'
		ORDER BY created_at DESC
	`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	scans := []gin.H{}
	for rows.Next() {
		var (
			id, repoName, planSummary, riskLevel, reasoning, status string
			riskScore                                               int
			createdAt                                               interface{}
		)
		if err := rows.Scan(&id, &repoName, &planSummary, &riskScore, &riskLevel, &reasoning, &status, &createdAt); err != nil {
			continue
		}
		scans = append(scans, gin.H{
			"id": id, "repo_name": repoName, "plan_summary": planSummary,
			"risk_score": riskScore, "risk_level": riskLevel, "reasoning": reasoning,
			"status": status, "created_at": createdAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{"scans": scans})
}

// --- Approvals ---

type decisionInput struct {
	Decision  string `json:"decision" binding:"required"` // "approved" | "rejected"
	DecidedBy string `json:"decided_by"`
}

// DecideScan records an approve/reject decision from the mobile app and
// updates the scan's status accordingly.
func (h *Handler) DecideScan(c *gin.Context) {
	scanID := c.Param("id")

	var input decisionInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if input.Decision != "approved" && input.Decision != "rejected" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "decision must be 'approved' or 'rejected'"})
		return
	}
	if input.DecidedBy == "" {
		input.DecidedBy = "on-call"
	}

	ctx := context.Background()
	tx, err := h.Pool.Begin(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer tx.Rollback(ctx)

	tag, err := tx.Exec(ctx,
		`UPDATE plan_scans SET status = $1, updated_at = now() WHERE id = $2 AND status = 'scored'`,
		input.Decision, scanID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if tag.RowsAffected() == 0 {
		c.JSON(http.StatusConflict, gin.H{"error": "scan not found or already decided"})
		return
	}

	if _, err := tx.Exec(ctx,
		`INSERT INTO approval_audit (plan_scan_id, decision, decided_by) VALUES ($1, $2, $3)`,
		scanID, input.Decision, input.DecidedBy,
	); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if err := tx.Commit(ctx); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"id": scanID, "status": input.Decision})
}

// --- Device Registration ---

type registerDeviceInput struct {
	DeviceToken string `json:"device_token" binding:"required"`
	OwnerName   string `json:"owner_name"`
}

// RegisterDevice stores an FCM device token so the API knows where to
// push approval alerts. Called once by the Flutter app on startup/login.
func (h *Handler) RegisterDevice(c *gin.Context) {
	var input registerDeviceInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if input.OwnerName == "" {
		input.OwnerName = "on-call"
	}

	_, err := h.Pool.Exec(c.Request.Context(),
		`INSERT INTO device_tokens (device_token, owner_name) VALUES ($1, $2)
		 ON CONFLICT (device_token) DO NOTHING`,
		input.DeviceToken, input.OwnerName,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"registered": true})
}

// --- Drift Events ---

// ListDriftEvents returns recent drift events detected by the K8s operator.
func (h *Handler) ListDriftEvents(c *gin.Context) {
	rows, err := h.Pool.Query(c.Request.Context(), `
		SELECT id, resource_kind, resource_name, namespace, diff, detected_at
		FROM drift_events
		ORDER BY detected_at DESC
		LIMIT 50
	`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	events := []gin.H{}
	for rows.Next() {
		var (
			id, kind, name, namespace, diff string
			detectedAt                      interface{}
		)
		if err := rows.Scan(&id, &kind, &name, &namespace, &diff, &detectedAt); err != nil {
			continue
		}
		events = append(events, gin.H{
			"id": id, "resource_kind": kind, "resource_name": name,
			"namespace": namespace, "diff": diff, "detected_at": detectedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{"events": events})
}
