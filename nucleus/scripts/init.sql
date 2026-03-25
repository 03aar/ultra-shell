-- NUCLEUS Database Schema
-- PostgreSQL 15

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Sessions table
CREATE TABLE IF NOT EXISTS sessions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    shell_pid INTEGER,
    start_time TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    end_time TIMESTAMPTZ,
    git_repo TEXT,
    cwd TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_sessions_start_time ON sessions(start_time DESC);

-- API Keys table
CREATE TABLE IF NOT EXISTS api_keys (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    key_hash VARCHAR(128) NOT NULL UNIQUE,
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_used_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_api_keys_hash ON api_keys(key_hash);

-- Insert dev API key (stored as SHA256 hash for production, but also
-- store the raw key for dev mode so the simple string check works)
INSERT INTO api_keys (name, key_hash) VALUES (
    'Development Key',
    encode(digest('dev-nucleus-key-local', 'sha256'), 'hex')
) ON CONFLICT (key_hash) DO NOTHING;

-- Also insert the raw key for dev-mode direct matching
INSERT INTO api_keys (name, key_hash) VALUES (
    'Development Key (raw)',
    'dev-nucleus-key-local'
) ON CONFLICT (key_hash) DO NOTHING;

-- Audit logs table
CREATE TABLE IF NOT EXISTS audit_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    session_id UUID REFERENCES sessions(id) ON DELETE SET NULL,
    execution_id UUID,
    action VARCHAR(50) NOT NULL,
    actor VARCHAR(255) NOT NULL DEFAULT 'system',
    details JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_audit_logs_session ON audit_logs(session_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_created ON audit_logs(created_at DESC);

-- Rollback events table
CREATE TABLE IF NOT EXISTS rollback_events (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    execution_id UUID NOT NULL,
    session_id UUID REFERENCES sessions(id) ON DELETE SET NULL,
    files_restored TEXT[] NOT NULL DEFAULT '{}',
    env_restored JSONB NOT NULL DEFAULT '{}',
    success BOOLEAN NOT NULL DEFAULT true,
    message TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_rollback_events_execution ON rollback_events(execution_id);
CREATE INDEX IF NOT EXISTS idx_rollback_events_created ON rollback_events(created_at DESC);

-- Grant permissions
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO nucleus;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO nucleus;
