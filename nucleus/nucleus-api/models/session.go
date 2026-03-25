package models

import "time"

type Session struct {
	ID               string            `json:"id" db:"id"`
	Name             string            `json:"name" db:"name"`
	ShellPID         uint32            `json:"shell_pid" db:"shell_pid"`
	StartTime        time.Time         `json:"start_time" db:"start_time"`
	EndTime          *time.Time        `json:"end_time" db:"end_time"`
	GitRepo          *string           `json:"git_repo" db:"git_repo"`
	Cwd              string            `json:"cwd" db:"cwd"`
	CreatedAt        time.Time         `json:"created_at" db:"created_at"`
	EnvSnapshot      map[string]string `json:"env_snapshot,omitempty"`
	ExecutionCount   int               `json:"execution_count,omitempty"`
	RiskWarningCount int               `json:"risk_warning_count,omitempty"`
}

type CreateSessionRequest struct {
	Name string `json:"name" binding:"required"`
}

type SessionReplayResponse struct {
	Session    Session         `json:"session"`
	Executions []ExecutionNode `json:"executions"`
}

type APIResponse struct {
	Data  interface{} `json:"data"`
	Meta  APIMeta     `json:"meta"`
	Error *string     `json:"error"`
}

type APIMeta struct {
	Timestamp string `json:"timestamp"`
	SessionID string `json:"session_id,omitempty"`
	Version   string `json:"version"`
}

type WSMessage struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}
