package handlers

import (
	"fmt"
	"net/http"
	"nucleus-api/models"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type SkillDef struct {
	Name        string       `json:"name"`
	Description string       `json:"description"`
	Parameters  []SkillParam `json:"parameters"`
	Steps       []SkillStep  `json:"steps"`
}

type SkillParam struct {
	Name        string `json:"name" yaml:"name"`
	Type        string `json:"type" yaml:"type"`
	Required    bool   `json:"required" yaml:"required"`
	Default     string `json:"default,omitempty" yaml:"default"`
	Description string `json:"description,omitempty" yaml:"description"`
}

type SkillStep struct {
	Command              string `json:"command" yaml:"command"`
	Description          string `json:"description" yaml:"description"`
	Risk                 string `json:"risk" yaml:"risk"`
	RequiresConfirmation bool   `json:"requires_confirmation,omitempty" yaml:"requires_confirmation"`
	Condition            string `json:"condition,omitempty" yaml:"condition"`
}

type RunSkillRequest struct {
	Params map[string]interface{} `json:"params"`
}

type SkillResult struct {
	Success bool             `json:"success"`
	Steps   []SkillStepResult `json:"steps"`
	Message string           `json:"message"`
}

type SkillStepResult struct {
	Command     string `json:"command"`
	Description string `json:"description"`
	ExitCode    int    `json:"exit_code"`
	Output      string `json:"output"`
	Skipped     bool   `json:"skipped"`
}

// Built-in skills loaded at startup
var builtinSkills = []SkillDef{
	{
		Name:        "git_cleanup",
		Description: "Clean up local git branches that have been merged into main/master",
		Parameters:  []SkillParam{},
		Steps: []SkillStep{
			{Command: "git fetch --prune", Description: "Fetch latest and prune deleted remote branches", Risk: "low"},
			{Command: "git branch --merged main 2>/dev/null || git branch --merged master", Description: "List branches merged into main", Risk: "low"},
		},
	},
	{
		Name:        "docker_cleanup",
		Description: "Remove stopped containers, unused images, volumes, and networks",
		Parameters:  []SkillParam{{Name: "aggressive", Type: "boolean", Required: false, Default: "false"}},
		Steps: []SkillStep{
			{Command: "docker container prune -f", Description: "Remove stopped containers", Risk: "low"},
			{Command: "docker image prune -f", Description: "Remove dangling images", Risk: "low"},
		},
	},
	{
		Name:        "project_setup",
		Description: "Set up a new project directory with standard structure",
		Parameters: []SkillParam{
			{Name: "project_name", Type: "string", Required: true},
			{Name: "type", Type: "string", Required: true, Description: "One of: node, python, rust, go"},
		},
		Steps: []SkillStep{
			{Command: "mkdir -p {project_name}", Description: "Create project directory", Risk: "low"},
			{Command: "cd {project_name} && git init", Description: "Initialize git", Risk: "low"},
		},
	},
	{
		Name:        "deploy_check",
		Description: "Run pre-deployment checks before any deployment",
		Parameters:  []SkillParam{},
		Steps: []SkillStep{
			{Command: "git status --porcelain", Description: "Check for uncommitted changes", Risk: "low"},
			{Command: "git rev-parse --abbrev-ref HEAD", Description: "Check current branch", Risk: "low"},
		},
	},
	{
		Name:        "env_audit",
		Description: "Audit environment variables for exposed secrets and security issues",
		Parameters:  []SkillParam{},
		Steps: []SkillStep{
			{Command: "env | grep -iE '(secret|token|key|password)' | sed 's/=.*/=***REDACTED***/' || echo 'No sensitive variables detected'", Description: "Scan for sensitive variables", Risk: "low"},
		},
	},
	{
		Name:        "process_debug",
		Description: "Debug a running or crashed process by gathering diagnostics",
		Parameters:  []SkillParam{{Name: "pid_or_name", Type: "string", Required: true}},
		Steps: []SkillStep{
			{Command: "ps aux | grep -i {pid_or_name} | grep -v grep", Description: "Find process info", Risk: "low"},
		},
	},
}

func GetSkills(c *gin.Context) {
	skills := make([]map[string]interface{}, 0)
	for _, s := range builtinSkills {
		skills = append(skills, map[string]interface{}{
			"name":        s.Name,
			"description": s.Description,
			"parameters":  s.Parameters,
			"source":      "builtin",
		})
	}

	// Also check for user skills on filesystem
	userSkillsDir := os.Getenv("NUCLEUS_SKILLS_DIR")
	if userSkillsDir == "" {
		userSkillsDir = "/etc/nucleus/skills"
	}
	if entries, err := os.ReadDir(userSkillsDir); err == nil {
		for _, entry := range entries {
			if filepath.Ext(entry.Name()) == ".yaml" || filepath.Ext(entry.Name()) == ".yml" {
				name := strings.TrimSuffix(entry.Name(), filepath.Ext(entry.Name()))
				skills = append(skills, map[string]interface{}{
					"name":        name,
					"description": "User-defined skill",
					"source":      "user",
				})
			}
		}
	}

	respondSuccess(c, skills, "")
}

func GetSkill(c *gin.Context) {
	name := c.Param("name")

	for _, s := range builtinSkills {
		if s.Name == name {
			respondSuccess(c, s, "")
			return
		}
	}

	errMsg := "Skill not found: " + name
	c.JSON(http.StatusNotFound, models.APIResponse{
		Data:  nil,
		Meta:  models.APIMeta{Timestamp: time.Now().UTC().Format(time.RFC3339), Version: "1.0"},
		Error: &errMsg,
	})
}

func RunSkill(c *gin.Context) {
	name := c.Param("name")

	var found *SkillDef
	for _, s := range builtinSkills {
		if s.Name == name {
			found = &s
			break
		}
	}

	if found == nil {
		errMsg := "Skill not found: " + name
		c.JSON(http.StatusNotFound, models.APIResponse{
			Data:  nil,
			Meta:  models.APIMeta{Timestamp: time.Now().UTC().Format(time.RFC3339), Version: "1.0"},
			Error: &errMsg,
		})
		return
	}

	var req RunSkillRequest
	c.ShouldBindJSON(&req)

	// Execute skill steps (simulated — in production would use nucleus-core)
	results := make([]SkillStepResult, 0)
	for _, step := range found.Steps {
		cmd := step.Command
		// Substitute parameters
		if req.Params != nil {
			for key, val := range req.Params {
				cmd = strings.ReplaceAll(cmd, "{"+key+"}", interfaceToString(val))
			}
		}

		results = append(results, SkillStepResult{
			Command:     cmd,
			Description: step.Description,
			ExitCode:    0,
			Output:      "Step completed: " + step.Description,
			Skipped:     false,
		})
	}

	result := SkillResult{
		Success: true,
		Steps:   results,
		Message: "Skill '" + name + "' completed successfully",
	}

	respondSuccess(c, result, "")
}

func interfaceToString(v interface{}) string {
	switch val := v.(type) {
	case string:
		return val
	case float64:
		if val == float64(int64(val)) {
			return strings.TrimSuffix(strings.TrimSuffix(
				fmt.Sprintf("%g", val), ".0"), ".")
		}
		return fmt.Sprintf("%g", val)
	case bool:
		if val {
			return "true"
		}
		return "false"
	default:
		return fmt.Sprintf("%v", v)
	}
}
