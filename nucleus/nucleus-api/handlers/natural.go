package handlers

import (
	"net/http"
	"nucleus-api/models"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type TranslateRequest struct {
	Input   string      `json:"input" binding:"required"`
	Context interface{} `json:"context"`
}

type TranslateResponse struct {
	Command     string `json:"command"`
	Explanation string `json:"explanation"`
	RiskLevel   string `json:"risk_level"`
}

// NaturalTranslate converts natural language to a shell command
func NaturalTranslate(c *gin.Context) {
	var req TranslateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errMsg := "Invalid request: input is required"
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Data:  nil,
			Meta:  models.APIMeta{Timestamp: time.Now().UTC().Format(time.RFC3339), Version: "1.0"},
			Error: &errMsg,
		})
		return
	}

	result := translateNaturalLanguage(req.Input)
	respondSuccess(c, result, "")
}

func translateNaturalLanguage(input string) TranslateResponse {
	q := strings.ToLower(strings.TrimSpace(input))

	// Pattern-based translation for common queries
	translations := []struct {
		patterns    []string
		command     string
		explanation string
		risk        string
	}{
		{[]string{"large file", "big file", "files larger"}, "find . -size +100M -type f", "Find all files larger than 100MB in the current directory", "low"},
		{[]string{"what changed", "recent change", "what's different"}, "git diff --stat HEAD~5", "Show file changes in the last 5 commits", "low"},
		{[]string{"who is using port", "what's on port", "port 3000"}, "lsof -i :3000 || ss -tlnp | grep 3000", "Show processes using port 3000", "low"},
		{[]string{"port 8080"}, "lsof -i :8080 || ss -tlnp | grep 8080", "Show processes using port 8080", "low"},
		{[]string{"listening port", "open port", "all port"}, "ss -tlnp", "Show all listening TCP ports", "low"},
		{[]string{"disk space", "disk usage", "storage"}, "df -h", "Show disk usage for all mounted filesystems", "low"},
		{[]string{"memory", "ram usage"}, "free -h", "Show system memory usage", "low"},
		{[]string{"cpu", "process", "running", "top process"}, "ps aux --sort=-%cpu | head -20", "Show top 20 processes by CPU usage", "low"},
		{[]string{"undo", "rollback", "revert last"}, "nuc rollback --last", "Rollback the most recent Nucleus command", "medium"},
		{[]string{"git status", "repo status"}, "git status", "Show current git repository status", "low"},
		{[]string{"git log", "commit history", "recent commit"}, "git log --oneline -20", "Show last 20 git commits", "low"},
		{[]string{"find log", "log file"}, "find . -name '*.log' -type f 2>/dev/null | head -20", "Find all log files in the current directory", "low"},
		{[]string{"network connection", "connections"}, "ss -tunp", "Show all network connections", "low"},
		{[]string{"docker container", "running container"}, "docker ps -a", "List all Docker containers", "low"},
		{[]string{"env", "environment variable"}, "env | sort | head -50", "Show environment variables (sorted)", "low"},
		{[]string{"delete", "remove", "clean up"}, "echo 'Please specify what to delete'", "Specify the target for deletion", "high"},
		{[]string{"install"}, "echo 'Please specify what to install'", "Specify the package to install", "medium"},
		{[]string{"deploy"}, "echo 'Please specify deployment target'", "Specify the deployment configuration", "high"},
		{[]string{"test", "run test"}, "test -f Makefile && make test || test -f package.json && npm test || echo 'No test command detected'", "Run tests using detected test runner", "low"},
		{[]string{"build"}, "test -f Makefile && make build || test -f package.json && npm run build || echo 'No build command detected'", "Run build using detected build tool", "low"},
		{[]string{"ip address", "my ip"}, "hostname -I 2>/dev/null || ifconfig | grep inet | head -5", "Show local IP addresses", "low"},
		{[]string{"system info", "os info"}, "uname -a && cat /etc/os-release 2>/dev/null | head -5", "Show operating system information", "low"},
	}

	for _, t := range translations {
		for _, pattern := range t.patterns {
			if strings.Contains(q, pattern) {
				return TranslateResponse{
					Command:     t.command,
					Explanation: t.explanation,
					RiskLevel:   t.risk,
				}
			}
		}
	}

	return TranslateResponse{
		Command:     "echo 'Could not translate. Try: nuc ask \"your question\"'",
		Explanation: "Unable to translate this query. Try being more specific.",
		RiskLevel:   "low",
	}
}
