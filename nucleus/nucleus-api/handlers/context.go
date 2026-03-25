package handlers

import (
	"nucleus-api/models"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

func GetContext(c *gin.Context) {
	ctx := buildContextState()
	respondSuccess(c, ctx, "")
}

func GetContextGraph(c *gin.Context) {
	executionsMu.RLock()
	defer executionsMu.RUnlock()

	sessionID := c.Query("session_id")
	nodes := make([]models.ExecutionNode, 0)
	for _, exec := range executions {
		if sessionID == "" || exec.SessionID == sessionID {
			nodes = append(nodes, exec)
		}
	}

	nodeIDs := make(map[string]bool)
	for _, n := range nodes {
		nodeIDs[n.ID] = true
	}

	filteredEdges := make([]models.Edge, 0)
	for _, edge := range edges {
		if nodeIDs[edge.From] || nodeIDs[edge.To] {
			filteredEdges = append(filteredEdges, edge)
		}
	}

	respondSuccess(c, models.GraphResponse{
		Nodes: nodes,
		Edges: filteredEdges,
	}, sessionID)
}

func buildContextState() models.ContextState {
	cwd, _ := os.Getwd()
	gitBranch := runCommand("git", "rev-parse", "--abbrev-ref", "HEAD")
	gitStatus := runCommand("git", "status", "--porcelain")

	processes := getTopProcesses()
	envVars := getFilteredEnvVars()
	diskUsage := getDiskUsage()
	openFiles := getOpenFilesCount()

	return models.ContextState{
		Cwd:              cwd,
		GitBranch:        strings.TrimSpace(gitBranch),
		GitStatus:        strings.TrimSpace(gitStatus),
		RunningProcesses: processes,
		EnvVars:          envVars,
		DiskUsage:        diskUsage,
		OpenFilesCount:   openFiles,
	}
}

func runCommand(name string, args ...string) string {
	cmd := exec.Command(name, args...)
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return string(out)
}

func getTopProcesses() []models.ProcessInfo {
	processes := make([]models.ProcessInfo, 0, 20)

	switch runtime.GOOS {
	case "linux":
		out := runCommand("ps", "aux", "--sort=-%cpu")
		processes = parsePsOutput(out)
	case "darwin":
		// macOS uses BSD ps without --sort
		out := runCommand("ps", "aux", "-r")
		processes = parsePsOutput(out)
	case "windows":
		out := runCommand("tasklist", "/FO", "CSV", "/NH")
		lines := strings.Split(out, "\n")
		for i, line := range lines {
			if i >= 20 || strings.TrimSpace(line) == "" {
				break
			}
			fields := strings.Split(line, ",")
			if len(fields) < 5 {
				continue
			}
			name := strings.Trim(fields[0], "\"")
			pid, _ := strconv.Atoi(strings.Trim(fields[1], "\""))
			mem := strings.Trim(fields[4], "\" K\r\n")
			memKB, _ := strconv.ParseFloat(strings.ReplaceAll(mem, ",", ""), 64)
			processes = append(processes, models.ProcessInfo{
				PID:     pid,
				Name:    name,
				CPU:     0,
				Memory:  memKB / 1024,
				Command: name,
			})
		}
	}

	return processes
}

func parsePsOutput(out string) []models.ProcessInfo {
	lines := strings.Split(out, "\n")
	processes := make([]models.ProcessInfo, 0, 20)

	for i, line := range lines {
		if i == 0 || strings.TrimSpace(line) == "" {
			continue
		}
		if len(processes) >= 20 {
			break
		}

		fields := strings.Fields(line)
		if len(fields) < 11 {
			continue
		}

		pid, _ := strconv.Atoi(fields[1])
		cpu, _ := strconv.ParseFloat(fields[2], 64)
		mem, _ := strconv.ParseFloat(fields[3], 64)
		command := strings.Join(fields[10:], " ")

		processes = append(processes, models.ProcessInfo{
			PID:     pid,
			Name:    fields[10],
			CPU:     cpu,
			Memory:  mem,
			Command: command,
		})
	}

	return processes
}

func getFilteredEnvVars() map[string]string {
	vars := make(map[string]string)
	secrets := map[string]bool{
		"SECRET": true, "PASSWORD": true, "TOKEN": true, "KEY": true,
		"PRIVATE": true, "CREDENTIAL": true, "AUTH": true,
	}

	for _, env := range os.Environ() {
		parts := strings.SplitN(env, "=", 2)
		if len(parts) != 2 {
			continue
		}
		name := parts[0]
		value := parts[1]

		isSensitive := false
		upperName := strings.ToUpper(name)
		for keyword := range secrets {
			if strings.Contains(upperName, keyword) {
				isSensitive = true
				break
			}
		}

		if isSensitive {
			vars[name] = "***MASKED***"
		} else {
			vars[name] = value
		}
	}
	return vars
}

func getOpenFilesCount() int {
	switch runtime.GOOS {
	case "linux":
		out := runCommand("sh", "-c", "ls /proc/self/fd 2>/dev/null | wc -l")
		if out != "" {
			n, _ := strconv.Atoi(strings.TrimSpace(out))
			return n
		}
	case "darwin":
		out := runCommand("sh", "-c", "lsof -p $$ 2>/dev/null | wc -l")
		if out != "" {
			n, _ := strconv.Atoi(strings.TrimSpace(out))
			return n
		}
	}
	return 0
}
