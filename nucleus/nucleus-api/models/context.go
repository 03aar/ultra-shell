package models

type ContextState struct {
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

type GraphResponse struct {
	Nodes []ExecutionNode `json:"nodes"`
	Edges []Edge          `json:"edges"`
}

type RollbackResponse struct {
	Success       bool              `json:"success"`
	FilesRestored []string          `json:"files_restored"`
	EnvRestored   map[string]string `json:"env_restored"`
	Message       string            `json:"message"`
}
