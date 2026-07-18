-- 001_init.sql
-- Core tables for plan scans and drift events

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS plan_scans (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    repo_name TEXT NOT NULL,
    plan_summary TEXT NOT NULL,
    risk_score INTEGER NOT NULL DEFAULT 0,
    risk_level TEXT NOT NULL DEFAULT 'low' CHECK (risk_level IN ('low', 'medium', 'high')),
    reasoning TEXT,
    status TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'scored', 'approved', 'rejected')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS drift_events (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    resource_kind TEXT NOT NULL,
    resource_name TEXT NOT NULL,
    namespace TEXT NOT NULL DEFAULT 'default',
    diff TEXT NOT NULL,
    detected_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_plan_scans_status ON plan_scans(status);
CREATE INDEX IF NOT EXISTS idx_drift_events_detected_at ON drift_events(detected_at DESC);