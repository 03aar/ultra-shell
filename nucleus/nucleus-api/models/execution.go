package models

import "time"

type ParsedCommand struct {
	Raw            string            `json:"raw"`
	Binary         string            `json:"binary"`
	Args           []string          `json:"args"`
	Flags          []string          `json:"flags"`
	Pipes          []PipeSegment     `json:"pipes"`
	Redirections   []Redirection     `json:"redirections"`
	EnvAssignments map[string]string `json:"env_assignments"`
	Category       string            `json:"category"`
	IsBackground   bool              `json:"is_background"`
	IsChained      bool              `json:"is_chained"`
}

type PipeSegment struct {
	Binary string   `json:"binary"`
	Args   []string `json:"args"`
	Flags  []string `json:"flags"`
}

type Redirection struct {
	Direction string `json:"direction"`
	Target    string `json:"target"`
}

type RiskFlag struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Level   string `json:"level"`
}

type ExecutionNode struct {
	ID               string            `json:"id"`
	Timestamp        time.Time         `json:"timestamp"`
	Command          ParsedCommand     `json:"command"`
	ExitCode         int               `json:"exit_code"`
	DurationMs       uint64            `json:"duration_ms"`
	Stdout           string            `json:"stdout"`
	Stderr           string            `json:"stderr"`
	FilesMutated     []string          `json:"files_mutated"`
	EnvChanges       map[string]string `json:"env_changes"`
	ProcessesSpawned []uint32          `json:"processes_spawned"`
	RiskFlags        []RiskFlag        `json:"risk_flags"`
	RollbackAvailable bool            `json:"rollback_available"`
	RolledBack       bool              `json:"rolled_back"`
	SessionID        string            `json:"session_id"`
}

type ExecutionListResponse struct {
	Executions []ExecutionNode `json:"executions"`
	Total      int             `json:"total"`
	SessionID  string          `json:"session_id"`
}

type ExecutionContextResponse struct {
	Execution       ExecutionNode   `json:"execution"`
	Edges           []Edge          `json:"edges"`
	RecentPrior     []ExecutionNode `json:"recent_prior"`
	EnvDiffSinceStart map[string]string `json:"env_diff_since_start"`
}

type Edge struct {
	From           string `json:"from"`
	To             string `json:"to"`
	DependencyType string `json:"dependency_type"`
}
