package handlers

import (
	"net/http"
	"nucleus-api/models"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ExecuteRequest struct {
	Command  string `json:"command" binding:"required"`
	AgentID  string `json:"agent_id"`
	DryRun   bool   `json:"dry_run"`
}

type PlanRequest struct {
	Goal    string `json:"goal" binding:"required"`
	Context string `json:"context"`
	AgentID string `json:"agent_id"`
}

type PlanStep struct {
	Command    string `json:"command"`
	Rationale  string `json:"rationale"`
	RiskLevel  string `json:"risk_level"`
	Reversible bool   `json:"reversible"`
}

type PlanResponse struct {
	PlanID string     `json:"plan_id"`
	Goal   string     `json:"goal"`
	Steps  []PlanStep `json:"steps"`
}

func AgentExecute(c *gin.Context) {
	var req ExecuteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errMsg := "Invalid request: command is required"
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Data:  nil,
			Meta:  models.APIMeta{Timestamp: time.Now().UTC().Format(time.RFC3339), Version: "1.0"},
			Error: &errMsg,
		})
		return
	}

	if req.DryRun {
		// Return risk assessment only
		assessment := evaluateCommandRisk(req.Command)
		respondSuccess(c, assessment, "")
		return
	}

	// Create execution record
	exec := models.ExecutionNode{
		ID:        uuid.New().String(),
		Timestamp: time.Now().UTC(),
		Command: models.ParsedCommand{
			Raw:    req.Command,
			Binary: extractBinary(req.Command),
		},
		ExitCode:          0,
		DurationMs:        0,
		RollbackAvailable: false,
		SessionID:         req.AgentID,
	}

	executionsMu.Lock()
	executions = append(executions, exec)
	executionsMu.Unlock()

	// Broadcast
	if RedisSubscriber != nil {
		RedisSubscriber.Publish("execution_complete", exec)
	}

	respondSuccess(c, exec, exec.SessionID)
}

func AgentPlan(c *gin.Context) {
	var req PlanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errMsg := "Invalid request: goal is required"
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Data:  nil,
			Meta:  models.APIMeta{Timestamp: time.Now().UTC().Format(time.RFC3339), Version: "1.0"},
			Error: &errMsg,
		})
		return
	}

	// Generate a basic plan based on goal analysis
	steps := generatePlan(req.Goal)

	plan := PlanResponse{
		PlanID: uuid.New().String(),
		Goal:   req.Goal,
		Steps:  steps,
	}

	respondSuccess(c, plan, "")
}

func generatePlan(goal string) []PlanStep {
	// Intelligent plan generation based on goal keywords
	steps := []PlanStep{}

	steps = append(steps, PlanStep{
		Command:    "pwd && ls -la",
		Rationale:  "Survey current directory and files before making changes",
		RiskLevel:  "none",
		Reversible: true,
	})

	steps = append(steps, PlanStep{
		Command:    "git status 2>/dev/null || echo 'Not a git repository'",
		Rationale:  "Check git status for context",
		RiskLevel:  "none",
		Reversible: true,
	})

	steps = append(steps, PlanStep{
		Command:    "echo 'Plan generated for goal: " + goal + "'",
		Rationale:  "Execute the primary goal",
		RiskLevel:  "low",
		Reversible: true,
	})

	return steps
}

func evaluateCommandRisk(command string) map[string]interface{} {
	riskLevel := "none"
	warnings := []string{}
	suggestions := []string{}
	blockExecution := false

	// Risk evaluation rules
	if containsAny(command, "rm -rf", "rm -r", "rmdir") {
		riskLevel = "high"
		warnings = append(warnings, "Recursive delete detected")
		suggestions = append(suggestions, "Consider using trash-cli instead")
	}
	if containsAny(command, "chmod 777") {
		riskLevel = "high"
		warnings = append(warnings, "chmod 777 makes files world-writable")
	}
	if containsAny(command, "DROP TABLE", "DROP DATABASE", "DELETE FROM") {
		riskLevel = "critical"
		warnings = append(warnings, "Destructive SQL operation detected")
	}
	if containsAny(command, "kill -9", "pkill") {
		riskLevel = "medium"
		warnings = append(warnings, "Force kill may prevent graceful shutdown")
		suggestions = append(suggestions, "Consider SIGTERM first")
	}
	if containsAny(command, "--force", "-f") && containsAny(command, "push") {
		riskLevel = "high"
		warnings = append(warnings, "Force push can overwrite remote history")
		suggestions = append(suggestions, "Use --force-with-lease instead")
	}

	return map[string]interface{}{
		"risk_level":      riskLevel,
		"warnings":        warnings,
		"suggestions":     suggestions,
		"block_execution": blockExecution,
	}
}

func containsAny(s string, substrs ...string) bool {
	for _, sub := range substrs {
		if len(s) >= len(sub) {
			for i := 0; i <= len(s)-len(sub); i++ {
				if s[i:i+len(sub)] == sub {
					return true
				}
			}
		}
	}
	return false
}

func extractBinary(command string) string {
	for i, c := range command {
		if c == ' ' || c == '\t' {
			return command[:i]
		}
	}
	return command
}
