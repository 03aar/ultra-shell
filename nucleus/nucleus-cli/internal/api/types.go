package api

import "time"

type Execution struct {
	ID          string    `json:"id"`
	Command     string    `json:"command"`
	Output      string    `json:"output"`
	ExitCode    int       `json:"exit_code"`
	RiskLevel   string    `json:"risk_level"`
	SessionID   string    `json:"session_id"`
	Duration    float64   `json:"duration"`
	Timestamp   time.Time `json:"timestamp"`
	Reversible  bool      `json:"reversible"`
	DryRun      bool      `json:"dry_run"`
	Explanation string    `json:"explanation"`
	Warnings    []string  `json:"warnings"`
	Suggestions []string  `json:"suggestions"`
}

type Context struct {
	WorkingDir  string            `json:"working_dir"`
	GitBranch   string            `json:"git_branch"`
	GitRepo     string            `json:"git_repo"`
	GitDirty    bool              `json:"git_dirty"`
	User        string            `json:"user"`
	Hostname    string            `json:"hostname"`
	Shell       string            `json:"shell"`
	SessionID   string            `json:"session_id"`
	Environment map[string]string `json:"environment"`
}

type GraphData struct {
	Format  string      `json:"format"`
	Content string      `json:"content"`
	Nodes   []GraphNode `json:"nodes,omitempty"`
	Edges   []GraphEdge `json:"edges,omitempty"`
}

type GraphNode struct {
	ID    string `json:"id"`
	Label string `json:"label"`
	Type  string `json:"type"`
}

type GraphEdge struct {
	Source string `json:"source"`
	Target string `json:"target"`
	Label  string `json:"label"`
}

type RollbackResult struct {
	ExecutionID string `json:"execution_id"`
	Command     string `json:"command"`
	UndoCommand string `json:"undo_command"`
	Success     bool   `json:"success"`
	Message     string `json:"message"`
}

type Session struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Commands  int       `json:"commands"`
	Active    bool      `json:"active"`
}

type Skill struct {
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Parameters  map[string]string `json:"parameters"`
	Category    string            `json:"category"`
}

type SkillResult struct {
	Skill   string `json:"skill"`
	Success bool   `json:"success"`
	Output  string `json:"output"`
	Message string `json:"message"`
}

type Plan struct {
	Goal  string     `json:"goal"`
	Steps []PlanStep `json:"steps"`
	Risk  string     `json:"risk"`
	Notes string     `json:"notes"`
}

type PlanStep struct {
	Order       int    `json:"order"`
	Command     string `json:"command"`
	Description string `json:"description"`
	RiskLevel   string `json:"risk_level"`
	Reversible  bool   `json:"reversible"`
}

type HealthStatus struct {
	Service string `json:"service"`
	Status  string `json:"status"`
	Latency string `json:"latency"`
	Message string `json:"message"`
}

type WatchEvent struct {
	Type      string    `json:"type"`
	Data      Execution `json:"data"`
	Timestamp time.Time `json:"timestamp"`
}

type APIError struct {
	StatusCode int    `json:"status_code"`
	Message    string `json:"message"`
	Detail     string `json:"detail"`
}

func (e *APIError) Error() string {
	if e.Detail != "" {
		return e.Message + ": " + e.Detail
	}
	return e.Message
}
