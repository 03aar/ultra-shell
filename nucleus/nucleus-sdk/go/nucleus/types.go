package nucleus

import (
	"encoding/json"
	"time"
)

type APIResponse struct {
	Data  json.RawMessage `json:"data"`
	Meta  APIMeta         `json:"meta"`
	Error *string         `json:"error"`
}

type APIMeta struct {
	Timestamp string `json:"timestamp"`
	SessionID string `json:"session_id,omitempty"`
	Version   string `json:"version"`
}

type GetExecutionsOpts struct {
	Limit     int
	Offset    int
	SessionID string
	Search    string
	Category  string
	RiskLevel string
}

type RiskFlag struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Level   string `json:"level"`
}

type ParsedCommand struct {
	Raw            string            `json:"raw"`
	Binary         string            `json:"binary"`
	Args           []string          `json:"args"`
	Flags          []string          `json:"flags"`
	Category       string            `json:"category"`
	EnvAssignments map[string]string `json:"env_assignments"`
	IsDestructive  bool              `json:"is_destructive"`
}

type FileMutation struct {
	Path       string `json:"path"`
	Operation  string `json:"operation"`
	SizeBefore *int64 `json:"size_before,omitempty"`
	SizeAfter  *int64 `json:"size_after,omitempty"`
}

type ExecutionNode struct {
	ID                string            `json:"id"`
	SessionID         string            `json:"session_id"`
	Timestamp         time.Time         `json:"timestamp"`
	Command           ParsedCommand     `json:"command"`
	ExitCode          int               `json:"exit_code"`
	DurationMs        uint64            `json:"duration_ms"`
	Stdout            string            `json:"stdout"`
	Stderr            string            `json:"stderr"`
	FilesMutated      []string          `json:"files_mutated"`
	EnvChanges        map[string]string `json:"env_changes"`
	ProcessesSpawned  []uint32          `json:"processes_spawned"`
	RiskFlags         []RiskFlag        `json:"risk_flags"`
	RollbackAvailable bool              `json:"rollback_available"`
	RolledBack        bool              `json:"rolled_back"`
}

type ExecutionResult struct {
	ExecutionNode
	RiskAssessment *RiskAssessment `json:"risk_assessment,omitempty"`
}

type ExecutionListResponse struct {
	Executions []ExecutionNode `json:"executions"`
	Total      int             `json:"total"`
	SessionID  string          `json:"session_id"`
}

type RiskAssessment struct {
	RiskLevel      string   `json:"risk_level"`
	Warnings       []string `json:"warnings"`
	Suggestions    []string `json:"suggestions"`
	BlockExecution bool     `json:"block_execution"`
}

type ContextSnapshot struct {
	Cwd              string            `json:"cwd"`
	GitBranch        string            `json:"git_branch"`
	GitStatus        string            `json:"git_status"`
	RunningProcesses []ProcessInfo     `json:"running_processes"`
	EnvVars          map[string]string `json:"env_vars"`
	DiskUsage        DiskUsage         `json:"disk_usage"`
	OpenFilesCount   int               `json:"open_files_count"`
}

type ProcessInfo struct {
	PID     int     `json:"pid"`
	Name    string  `json:"name"`
	CPU     float64 `json:"cpu"`
	Memory  float64 `json:"memory"`
	Command string  `json:"command"`
}

type DiskUsage struct {
	Total     uint64  `json:"total"`
	Used      uint64  `json:"used"`
	Available uint64  `json:"available"`
	Percent   float64 `json:"percent"`
}

type Edge struct {
	From           string `json:"from"`
	To             string `json:"to"`
	DependencyType string `json:"dependency_type"`
}

type GraphData struct {
	Nodes []ExecutionNode `json:"nodes"`
	Edges []Edge          `json:"edges"`
}

type RollbackResult struct {
	Success       bool              `json:"success"`
	FilesRestored []string          `json:"files_restored"`
	EnvRestored   map[string]string `json:"env_restored"`
	Message       string            `json:"message"`
}

type Session struct {
	ID               string     `json:"id"`
	Name             string     `json:"name"`
	ShellPID         uint32     `json:"shell_pid"`
	StartTime        time.Time  `json:"start_time"`
	EndTime          *time.Time `json:"end_time,omitempty"`
	GitRepo          *string    `json:"git_repo,omitempty"`
	Cwd              string     `json:"cwd"`
	CommandCount     int        `json:"command_count,omitempty"`
	RiskCount        int        `json:"risk_warning_count,omitempty"`
	ExecutionCount   int        `json:"execution_count,omitempty"`
	Tags             []string   `json:"tags,omitempty"`
}

type SessionReplay struct {
	Session    Session         `json:"session"`
	Executions []ExecutionNode `json:"executions"`
}

type PlanStep struct {
	Command    string `json:"command"`
	Rationale  string `json:"rationale"`
	RiskLevel  string `json:"risk_level"`
	Reversible bool   `json:"reversible"`
}

type PlanResponse struct {
	Steps []PlanStep `json:"steps"`
	Goal  string     `json:"goal"`
}

type Skill struct {
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Parameters  []SkillParam      `json:"parameters"`
}

type SkillParam struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Required bool   `json:"required"`
	Default  string `json:"default,omitempty"`
}

type SkillResult struct {
	Success bool            `json:"success"`
	Steps   []SkillStepResult `json:"steps"`
	Message string          `json:"message"`
}

type SkillStepResult struct {
	Command  string `json:"command"`
	ExitCode int    `json:"exit_code"`
	Output   string `json:"output"`
}
