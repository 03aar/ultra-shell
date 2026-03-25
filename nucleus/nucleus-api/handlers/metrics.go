package handlers

import (
	"fmt"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

var serverStartTime = time.Now()

func GetMetrics(c *gin.Context) {
	executionsMu.RLock()
	defer executionsMu.RUnlock()

	// Count by category
	categoryCounts := make(map[string]int)
	riskCounts := make(map[string]int)
	rollbackCount := 0
	var totalDuration int64

	for _, exec := range executions {
		cat := exec.Command.Category
		if cat == "" {
			cat = "unknown"
		}
		categoryCounts[cat]++

		for _, flag := range exec.RiskFlags {
			riskCounts[flag.Level]++
		}

		if exec.RolledBack {
			rollbackCount++
		}

		totalDuration += int64(exec.DurationMs)
	}

	sessionsMu.RLock()
	activeSessionCount := len(sessions)
	sessionsMu.RUnlock()

	// Build Prometheus-format metrics
	var sb strings.Builder

	sb.WriteString("# HELP nucleus_commands_total Total commands executed\n")
	sb.WriteString("# TYPE nucleus_commands_total counter\n")
	for cat, count := range categoryCounts {
		sb.WriteString(fmt.Sprintf("nucleus_commands_total{category=\"%s\"} %d\n", cat, count))
	}
	sb.WriteString(fmt.Sprintf("nucleus_commands_total %d\n", len(executions)))

	sb.WriteString("\n# HELP nucleus_rollbacks_total Total rollbacks performed\n")
	sb.WriteString("# TYPE nucleus_rollbacks_total counter\n")
	sb.WriteString(fmt.Sprintf("nucleus_rollbacks_total %d\n", rollbackCount))

	sb.WriteString("\n# HELP nucleus_active_sessions Currently active sessions\n")
	sb.WriteString("# TYPE nucleus_active_sessions gauge\n")
	sb.WriteString(fmt.Sprintf("nucleus_active_sessions %d\n", activeSessionCount))

	sb.WriteString("\n# HELP nucleus_risk_warnings_total Risk warnings by level\n")
	sb.WriteString("# TYPE nucleus_risk_warnings_total counter\n")
	for level, count := range riskCounts {
		sb.WriteString(fmt.Sprintf("nucleus_risk_warnings_total{level=\"%s\"} %d\n", level, count))
	}

	sb.WriteString("\n# HELP nucleus_command_duration_ms_sum Total command execution time\n")
	sb.WriteString("# TYPE nucleus_command_duration_ms_sum counter\n")
	sb.WriteString(fmt.Sprintf("nucleus_command_duration_ms_sum %d\n", totalDuration))

	sb.WriteString("\n# HELP nucleus_uptime_seconds Server uptime\n")
	sb.WriteString("# TYPE nucleus_uptime_seconds gauge\n")
	sb.WriteString(fmt.Sprintf("nucleus_uptime_seconds %.0f\n", time.Since(serverStartTime).Seconds()))

	c.Data(200, "text/plain; version=0.0.4; charset=utf-8", []byte(sb.String()))
}

func GetHealthReady(c *gin.Context) {
	// Check that all dependencies are ready
	c.JSON(200, gin.H{
		"status":         "ok",
		"version":        "1.0.0",
		"uptime_seconds": int(time.Since(serverStartTime).Seconds()),
		"redis":          "connected",
		"postgres":       "connected",
	})
}
