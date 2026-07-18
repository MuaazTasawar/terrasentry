package db

import "time"

// RiskLevel mirrors the risk tiers produced by the risk-scoring service.
type RiskLevel string

const (
	RiskLow    RiskLevel = "low"
	RiskMedium RiskLevel = "medium"
	RiskHigh   RiskLevel = "high"
)

// ScanStatus tracks the lifecycle of a Terraform plan scan.
type ScanStatus string

const (
	ScanPending  ScanStatus = "pending"
	ScanScored   ScanStatus = "scored"
	ScanApproved ScanStatus = "approved"
	ScanRejected ScanStatus = "rejected"
)

// PlanScan represents a single Terraform plan submitted for risk analysis.
type PlanScan struct {
	ID          string     `json:"id" db:"id"`
	RepoName    string     `json:"repo_name" db:"repo_name"`
	PlanSummary string     `json:"plan_summary" db:"plan_summary"`
	RiskScore   int        `json:"risk_score" db:"risk_score"`
	RiskLevel   RiskLevel  `json:"risk_level" db:"risk_level"`
	Reasoning   string     `json:"reasoning" db:"reasoning"`
	Status      ScanStatus `json:"status" db:"status"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
}

// DriftEvent represents a detected divergence between desired and live cluster state.
type DriftEvent struct {
	ID           string    `json:"id" db:"id"`
	ResourceKind string    `json:"resource_kind" db:"resource_kind"`
	ResourceName string    `json:"resource_name" db:"resource_name"`
	Namespace    string    `json:"namespace" db:"namespace"`
	Diff         string    `json:"diff" db:"diff"`
	DetectedAt   time.Time `json:"detected_at" db:"detected_at"`
}
