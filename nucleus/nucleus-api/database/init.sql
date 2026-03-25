-- NUCLEUS Database Schema — Full Production
-- PostgreSQL 15

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Sessions
CREATE TABLE IF NOT EXISTS sessions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    shell_binary VARCHAR(100),
    shell_pid INTEGER,
    cwd TEXT,
    git_repo TEXT,
    start_time TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    end_time TIMESTAMPTZ,
    command_count INTEGER DEFAULT 0,
    risk_count INTEGER DEFAULT 0,
    rollback_count INTEGER DEFAULT 0,
    tags TEXT[] DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_sessions_start ON sessions(start_time DESC);

-- Executions
CREATE TABLE IF NOT EXISTS executions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    session_id UUID REFERENCES sessions(id) ON DELETE SET NULL,
    sequence BIGINT,
    raw_command TEXT NOT NULL,
    binary_name VARCHAR(255),
    category VARCHAR(50),
    exit_code INTEGER,
    duration_ms BIGINT,
    stdout TEXT,
    stderr TEXT,
    risk_level VARCHAR(20) DEFAULT 'none',
    risk_warnings JSONB DEFAULT '[]',
    files_mutated JSONB DEFAULT '[]',
    env_changes JSONB DEFAULT '{}',
    rollback_available BOOLEAN DEFAULT FALSE,
    rolled_back BOOLEAN DEFAULT FALSE,
    timestamp TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_executions_session ON executions(session_id);
CREATE INDEX IF NOT EXISTS idx_executions_timestamp ON executions(timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_executions_category ON executions(category);
CREATE INDEX IF NOT EXISTS idx_executions_risk ON executions(risk_level);

-- API Keys
CREATE TABLE IF NOT EXISTS api_keys (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    key_hash VARCHAR(128) NOT NULL UNIQUE,
    scopes TEXT[] DEFAULT ARRAY['*'],
    last_used_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_api_keys_hash ON api_keys(key_hash);

-- Seed dev API key
INSERT INTO api_keys (name, key_hash, scopes) VALUES (
    'Development Key',
    'dev-nucleus-key-local',
    ARRAY['*']
) ON CONFLICT (key_hash) DO NOTHING;

-- Audit logs
CREATE TABLE IF NOT EXISTS audit_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    session_id UUID REFERENCES sessions(id) ON DELETE SET NULL,
    execution_id UUID,
    action VARCHAR(50) NOT NULL,
    actor VARCHAR(255) NOT NULL DEFAULT 'system',
    details JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_audit_session ON audit_logs(session_id);
CREATE INDEX IF NOT EXISTS idx_audit_created ON audit_logs(created_at DESC);

-- Rollback events
CREATE TABLE IF NOT EXISTS rollback_events (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    execution_id UUID REFERENCES executions(id) ON DELETE SET NULL,
    session_id UUID REFERENCES sessions(id) ON DELETE SET NULL,
    files_restored JSONB DEFAULT '[]',
    env_restored JSONB DEFAULT '{}',
    triggered_by VARCHAR(50) DEFAULT 'user',
    success BOOLEAN DEFAULT TRUE,
    message TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_rollback_exec ON rollback_events(execution_id);

-- Skills
CREATE TABLE IF NOT EXISTS skills (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) UNIQUE NOT NULL,
    description TEXT,
    source_path TEXT,
    parameters JSONB DEFAULT '[]',
    steps JSONB DEFAULT '[]',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Agent sessions
CREATE TABLE IF NOT EXISTS agent_sessions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    agent_id VARCHAR(100),
    provider VARCHAR(50),
    model VARCHAR(100),
    nucleus_session_id UUID REFERENCES sessions(id) ON DELETE SET NULL,
    goal TEXT,
    commands_executed INTEGER DEFAULT 0,
    status VARCHAR(20) DEFAULT 'active',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Grants
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO nucleus;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO nucleus;
