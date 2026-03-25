package nucleus

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

// NucleusClient is the Go SDK client for the Nucleus API.
type NucleusClient struct {
	APIUrl     string
	APIKey     string
	HTTPClient *http.Client
}

// NewClient creates a new Nucleus API client.
func NewClient(apiURL, apiKey string) *NucleusClient {
	return &NucleusClient{
		APIUrl:     apiURL + "/api/v1",
		APIKey:     apiKey,
		HTTPClient: &http.Client{Timeout: 30 * time.Second},
	}
}

// NewDefaultClient creates a client with default localhost settings.
func NewDefaultClient() *NucleusClient {
	return NewClient("http://localhost:8080", "dev-nucleus-key-local")
}

func (c *NucleusClient) doRequest(method, path string, body interface{}) (json.RawMessage, error) {
	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshal body: %w", err)
		}
		bodyReader = bytes.NewReader(data)
	}

	req, err := http.NewRequest(method, c.APIUrl+path, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Nucleus-Key", c.APIKey)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	var apiResp APIResponse
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}

	if apiResp.Error != nil {
		return nil, fmt.Errorf("API error: %s", *apiResp.Error)
	}

	return apiResp.Data, nil
}

// Execute runs a command through Nucleus.
func (c *NucleusClient) Execute(command string, dryRun bool) (*ExecutionResult, error) {
	body := map[string]interface{}{
		"command": command,
		"dry_run": dryRun,
	}
	data, err := c.doRequest("POST", "/agent/execute", body)
	if err != nil {
		return nil, err
	}
	var result ExecutionResult
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetContext returns the current environment state.
func (c *NucleusClient) GetContext() (*ContextSnapshot, error) {
	data, err := c.doRequest("GET", "/context", nil)
	if err != nil {
		return nil, err
	}
	var ctx ContextSnapshot
	if err := json.Unmarshal(data, &ctx); err != nil {
		return nil, err
	}
	return &ctx, nil
}

// GetExecutions returns recent executions.
func (c *NucleusClient) GetExecutions(opts *GetExecutionsOpts) ([]ExecutionNode, error) {
	params := url.Values{}
	if opts != nil {
		if opts.Limit > 0 {
			params.Set("limit", strconv.Itoa(opts.Limit))
		}
		if opts.SessionID != "" {
			params.Set("session_id", opts.SessionID)
		}
		if opts.Search != "" {
			params.Set("search", opts.Search)
		}
		if opts.Category != "" {
			params.Set("command_category", opts.Category)
		}
		if opts.RiskLevel != "" {
			params.Set("risk_level", opts.RiskLevel)
		}
	}
	qs := ""
	if len(params) > 0 {
		qs = "?" + params.Encode()
	}
	data, err := c.doRequest("GET", "/executions"+qs, nil)
	if err != nil {
		return nil, err
	}
	var resp ExecutionListResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}
	return resp.Executions, nil
}

// GetExecution returns a single execution by ID.
func (c *NucleusClient) GetExecution(id string) (*ExecutionNode, error) {
	data, err := c.doRequest("GET", "/executions/"+id, nil)
	if err != nil {
		return nil, err
	}
	var node ExecutionNode
	if err := json.Unmarshal(data, &node); err != nil {
		return nil, err
	}
	return &node, nil
}

// Rollback rolls back an execution.
func (c *NucleusClient) Rollback(executionID string) (*RollbackResult, error) {
	data, err := c.doRequest("POST", "/rollback/"+executionID, nil)
	if err != nil {
		return nil, err
	}
	var result RollbackResult
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetGraph returns the execution DAG.
func (c *NucleusClient) GetGraph(sessionID string) (*GraphData, error) {
	qs := ""
	if sessionID != "" {
		qs = "?session_id=" + sessionID
	}
	data, err := c.doRequest("GET", "/context/graph"+qs, nil)
	if err != nil {
		return nil, err
	}
	var graph GraphData
	if err := json.Unmarshal(data, &graph); err != nil {
		return nil, err
	}
	return &graph, nil
}

// GetSessions returns all sessions.
func (c *NucleusClient) GetSessions() ([]Session, error) {
	data, err := c.doRequest("GET", "/sessions", nil)
	if err != nil {
		return nil, err
	}
	var sessions []Session
	if err := json.Unmarshal(data, &sessions); err != nil {
		return nil, err
	}
	return sessions, nil
}

// CreateSession creates a new named session.
func (c *NucleusClient) CreateSession(name string) (*Session, error) {
	data, err := c.doRequest("POST", "/sessions", map[string]string{"name": name})
	if err != nil {
		return nil, err
	}
	var session Session
	if err := json.Unmarshal(data, &session); err != nil {
		return nil, err
	}
	return &session, nil
}

// GetSessionReplay returns all executions for replay.
func (c *NucleusClient) GetSessionReplay(sessionID string) (*SessionReplay, error) {
	data, err := c.doRequest("GET", "/sessions/"+sessionID+"/replay", nil)
	if err != nil {
		return nil, err
	}
	var replay SessionReplay
	if err := json.Unmarshal(data, &replay); err != nil {
		return nil, err
	}
	return &replay, nil
}

// Plan generates a command plan for a goal.
func (c *NucleusClient) Plan(goal string) (*PlanResponse, error) {
	data, err := c.doRequest("POST", "/agent/plan", map[string]interface{}{
		"goal":    goal,
		"context": "",
	})
	if err != nil {
		return nil, err
	}
	var plan PlanResponse
	if err := json.Unmarshal(data, &plan); err != nil {
		return nil, err
	}
	return &plan, nil
}

// GetSkills returns all available skills.
func (c *NucleusClient) GetSkills() ([]Skill, error) {
	data, err := c.doRequest("GET", "/skills", nil)
	if err != nil {
		return nil, err
	}
	var skills []Skill
	if err := json.Unmarshal(data, &skills); err != nil {
		return nil, err
	}
	return skills, nil
}

// RunSkill executes a skill.
func (c *NucleusClient) RunSkill(name string, params map[string]interface{}) (*SkillResult, error) {
	data, err := c.doRequest("POST", "/skills/"+name+"/run", map[string]interface{}{
		"params": params,
	})
	if err != nil {
		return nil, err
	}
	var result SkillResult
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// HealthCheck verifies the API is reachable.
func (c *NucleusClient) HealthCheck() error {
	req, err := http.NewRequest("GET", c.APIUrl[:len(c.APIUrl)-7]+"/health", nil)
	if err != nil {
		return err
	}
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("API unreachable: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf("API returned status %d", resp.StatusCode)
	}
	return nil
}
