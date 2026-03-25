---
sidebar_position: 3
title: Go SDK
---

# Go SDK

The Nucleus Go SDK provides a typed client for Go applications.

## Installation

```bash
go get github.com/03aar/ultra-shell/nucleus-sdk/go
```

Or from the monorepo:

```bash
cd nucleus-sdk/go
go build ./...
```

## Quick Start

```go
package main

import (
    "fmt"
    "log"

    nucleus "github.com/03aar/ultra-shell/nucleus-sdk/go"
)

func main() {
    client := nucleus.NewClient(
        nucleus.WithAPIURL("http://localhost:8080"),
        nucleus.WithAPIKey("dev-nucleus-key-local"),
    )

    // Execute a command
    result, err := client.Execute("ls -la")
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Exit: %d\n", result.ExitCode)
    fmt.Printf("Output: %s\n", result.Stdout)

    // Get context
    ctx, err := client.GetContext()
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("CWD: %s\n", ctx.WorkingDirectory)
    fmt.Printf("Branch: %s\n", ctx.Git.Branch)
}
```

## API Reference

### Client Construction

```go
// With options
client := nucleus.NewClient(
    nucleus.WithAPIURL("http://localhost:8080"),
    nucleus.WithAPIKey("dev-nucleus-key-local"),
    nucleus.WithTimeout(30 * time.Second),
)

// From environment variables
// Uses NUCLEUS_API_URL and NUCLEUS_API_KEY
client := nucleus.NewClientFromEnv()
```

### `client.Execute(command string, opts ...ExecuteOption) (*ExecutionResult, error)`

```go
// Basic execution
result, err := client.Execute("npm test")

// With options
result, err := client.Execute("rm -rf /tmp/data",
    nucleus.WithDryRun(true),
    nucleus.WithAgentID("my-agent"),
)

// Result fields
type ExecutionResult struct {
    ID                string        `json:"id"`
    Command           ParsedCommand `json:"command"`
    ExitCode          int           `json:"exit_code"`
    Stdout            string        `json:"stdout"`
    Stderr            string        `json:"stderr"`
    DurationMs        int64         `json:"duration_ms"`
    RiskLevel         string        `json:"risk_level"`
    RiskFlags         []string      `json:"risk_flags"`
    RollbackAvailable bool          `json:"rollback_available"`
    SessionID         string        `json:"session_id"`
    Timestamp         time.Time     `json:"timestamp"`
}
```

### `client.GetContext() (*Context, error)`

```go
ctx, err := client.GetContext()

type Context struct {
    WorkingDirectory string            `json:"working_directory"`
    Shell            string            `json:"shell"`
    User             string            `json:"user"`
    Hostname         string            `json:"hostname"`
    Git              GitContext         `json:"git"`
    Environment      map[string]string `json:"environment"`
    Processes        []Process         `json:"processes"`
}

type GitContext struct {
    IsRepo        bool     `json:"is_repo"`
    Branch        string   `json:"branch"`
    RepoRoot      string   `json:"repo_root"`
    ModifiedFiles []string `json:"modified_files"`
    StagedFiles   []string `json:"staged_files"`
    Ahead         int      `json:"ahead"`
    Behind        int      `json:"behind"`
}
```

### `client.GetHistory(opts ...HistoryOption) ([]ExecutionResult, error)`

```go
history, err := client.GetHistory(
    nucleus.WithLimit(10),
    nucleus.WithSearch("docker"),
)
```

### `client.Rollback(executionID string) (*RollbackResult, error)`

```go
result, err := client.Rollback("550e8400-...")

type RollbackResult struct {
    ExecutionID   string         `json:"execution_id"`
    FilesRestored []RestoredFile `json:"files_restored"`
    RolledBackAt  time.Time      `json:"rolled_back_at"`
}
```

### `client.CreatePlan(goal string, context string) (*Plan, error)`

```go
plan, err := client.CreatePlan("Deploy to staging", "")
for _, step := range plan.Steps {
    fmt.Printf("  %s [%s]\n", step.Command, step.RiskLevel)
}
```

### `client.RunSkill(name string, params map[string]string) (*SkillResult, error)`

```go
result, err := client.RunSkill("git_cleanup", nil)
result, err = client.RunSkill("project_setup", map[string]string{
    "project_name": "myapp",
    "type":         "node",
})
```

## WebSocket Streaming

```go
ch, err := client.Watch([]string{"executions", "warnings"})
if err != nil {
    log.Fatal(err)
}

for event := range ch {
    switch event.Type {
    case "execution":
        fmt.Printf("[EXEC] %s -> %d\n", event.Execution.Command, event.Execution.ExitCode)
    case "warning":
        fmt.Printf("[WARN] %s risk:%s\n", event.Warning.Command, event.Warning.RiskLevel)
    }
}
```

## Error Handling

```go
result, err := client.Execute("rm -rf /")
if err != nil {
    var blocked *nucleus.CommandBlockedError
    if errors.As(err, &blocked) {
        fmt.Printf("Blocked: risk=%s flags=%v\n", blocked.RiskLevel, blocked.RiskFlags)
    }

    var apiErr *nucleus.APIError
    if errors.As(err, &apiErr) {
        fmt.Printf("API error: %s %s\n", apiErr.Code, apiErr.Message)
    }
}
```

## Context-Aware Client

Use Go's `context.Context` for cancellation and timeouts:

```go
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()

result, err := client.ExecuteWithContext(ctx, "long-running-command")
```
