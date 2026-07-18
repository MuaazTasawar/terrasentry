-- 002_approvals.sql
-- Device registration for push notifications, and an audit trail of who
-- approved/rejected which scan and when.

CREATE TABLE IF NOT EXISTS device_tokens (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    device_token TEXT NOT NULL UNIQUE,
    owner_name TEXT NOT NULL DEFAULT 'on-call',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS approval_audit (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    plan_scan_id UUID NOT NULL REFERENCES plan_scans(id) ON DELETE CASCADE,
    decision TEXT NOT NULL CHECK (decision IN ('approved', 'rejected')),
    decided_by TEXT NOT NULL DEFAULT 'on-call',
    decided_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_approval_audit_plan_scan_id ON approval_audit(plan_scan_id);