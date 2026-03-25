package api

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

type NucleusClient struct {
	apiURL     string
	apiKey     string
	httpClient *http.Client
}

func NewClient(apiURL, apiKey string) *NucleusClient {
	return &NucleusClient{
		apiURL: apiURL,
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *NucleusClient) doRequest(method, path string, body interface{}) ([]byte, error) {
	var reqBody io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewReader(data)
	}

	reqURL := c.apiURL + path
	req, err := http.NewRequest(method, reqURL, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	if c.apiKey != "" {
		req.Header.Set("X-Nucleus-Key", c.apiKey)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode >= 400 {
		apiErr := &APIError{StatusCode: resp.StatusCode}
		if jsonErr := json.Unmarshal(respBody, apiErr); jsonErr != nil {
			apiErr.Message = fmt.Sprintf("HTTP %d: %s", resp.StatusCode, string(respBody))
		}
		return nil, apiErr
	}

	return respBody, nil
}

// PostRaw sends raw JSON bytes and returns the parsed data field
func (c *NucleusClient) PostRaw(path string, jsonBody []byte) (interface{}, error) {
	reqURL := c.apiURL + path
	req, err := http.NewRequest("POST", reqURL, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Nucleus-Key", c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var apiResp struct {
		Data  interface{} `json:"data"`
		Error *string     `json:"error"`
	}
	json.Unmarshal(body, &apiResp)
	if apiResp.Error != nil {
		return nil, fmt.Errorf("%s", *apiResp.Error)
	}
	return apiResp.Data, nil
}

func (c *NucleusClient) GetExecutions(limit int, search, sessionID string) ([]Execution, error) {
	params := url.Values{}
	if limit > 0 {
		params.Set("limit", strconv.Itoa(limit))
	}
	if search != "" {
		params.Set("search", search)
	}
	if sessionID != "" {
		params.Set("session_id", sessionID)
	}

	path := "/api/executions"
	if len(params) > 0 {
		path += "?" + params.Encode()
	}

	data, err := c.doRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}

	var executions []Execution
	if err := json.Unmarshal(data, &executions); err != nil {
		return nil, fmt.Errorf("failed to parse executions: %w", err)
	}
	return executions, nil
}

func (c *NucleusClient) GetContext() (Context, error) {
	data, err := c.doRequest("GET", "/api/context", nil)
	if err != nil {
		return Context{}, err
	}

	var ctx Context
	if err := json.Unmarshal(data, &ctx); err != nil {
		return Context{}, fmt.Errorf("failed to parse context: %w", err)
	}
	return ctx, nil
}

func (c *NucleusClient) GetGraph(sessionID, format string) (GraphData, error) {
	params := url.Values{}
	if sessionID != "" {
		params.Set("session_id", sessionID)
	}
	if format != "" {
		params.Set("format", format)
	}

	path := "/api/graph"
	if len(params) > 0 {
		path += "?" + params.Encode()
	}

	data, err := c.doRequest("GET", path, nil)
	if err != nil {
		return GraphData{}, err
	}

	var graph GraphData
	if err := json.Unmarshal(data, &graph); err != nil {
		return GraphData{}, fmt.Errorf("failed to parse graph: %w", err)
	}
	return graph, nil
}

func (c *NucleusClient) Execute(command string, dryRun bool) (Execution, error) {
	payload := map[string]interface{}{
		"command": command,
		"dry_run": dryRun,
	}

	data, err := c.doRequest("POST", "/api/execute", payload)
	if err != nil {
		return Execution{}, err
	}

	var exec Execution
	if err := json.Unmarshal(data, &exec); err != nil {
		return Execution{}, fmt.Errorf("failed to parse execution: %w", err)
	}
	return exec, nil
}

func (c *NucleusClient) Rollback(executionID string) (RollbackResult, error) {
	payload := map[string]interface{}{
		"execution_id": executionID,
	}

	data, err := c.doRequest("POST", "/api/rollback", payload)
	if err != nil {
		return RollbackResult{}, err
	}

	var result RollbackResult
	if err := json.Unmarshal(data, &result); err != nil {
		return RollbackResult{}, fmt.Errorf("failed to parse rollback result: %w", err)
	}
	return result, nil
}

func (c *NucleusClient) GetSessions() ([]Session, error) {
	data, err := c.doRequest("GET", "/api/sessions", nil)
	if err != nil {
		return nil, err
	}

	var sessions []Session
	if err := json.Unmarshal(data, &sessions); err != nil {
		return nil, fmt.Errorf("failed to parse sessions: %w", err)
	}
	return sessions, nil
}

func (c *NucleusClient) CreateSession(name string) (Session, error) {
	payload := map[string]interface{}{
		"name": name,
	}

	data, err := c.doRequest("POST", "/api/sessions", payload)
	if err != nil {
		return Session{}, err
	}

	var session Session
	if err := json.Unmarshal(data, &session); err != nil {
		return Session{}, fmt.Errorf("failed to parse session: %w", err)
	}
	return session, nil
}

func (c *NucleusClient) GetSessionReplay(id string) ([]Execution, error) {
	data, err := c.doRequest("GET", "/api/sessions/"+id+"/replay", nil)
	if err != nil {
		return nil, err
	}

	var executions []Execution
	if err := json.Unmarshal(data, &executions); err != nil {
		return nil, fmt.Errorf("failed to parse session replay: %w", err)
	}
	return executions, nil
}

func (c *NucleusClient) RunSkill(name string, params map[string]string) (SkillResult, error) {
	payload := map[string]interface{}{
		"name":   name,
		"params": params,
	}

	data, err := c.doRequest("POST", "/api/skills/run", payload)
	if err != nil {
		return SkillResult{}, err
	}

	var result SkillResult
	if err := json.Unmarshal(data, &result); err != nil {
		return SkillResult{}, fmt.Errorf("failed to parse skill result: %w", err)
	}
	return result, nil
}

func (c *NucleusClient) GetSkills() ([]Skill, error) {
	data, err := c.doRequest("GET", "/api/skills", nil)
	if err != nil {
		return nil, err
	}

	var skills []Skill
	if err := json.Unmarshal(data, &skills); err != nil {
		return nil, fmt.Errorf("failed to parse skills: %w", err)
	}
	return skills, nil
}

func (c *NucleusClient) Plan(goal string) (Plan, error) {
	payload := map[string]interface{}{
		"goal": goal,
	}

	data, err := c.doRequest("POST", "/api/plan", payload)
	if err != nil {
		return Plan{}, err
	}

	var plan Plan
	if err := json.Unmarshal(data, &plan); err != nil {
		return Plan{}, fmt.Errorf("failed to parse plan: %w", err)
	}
	return plan, nil
}

func (c *NucleusClient) HealthCheck() error {
	_, err := c.doRequest("GET", "/api/health", nil)
	return err
}

func (c *NucleusClient) WebSocketURL() string {
	wsURL := c.apiURL
	if len(wsURL) > 4 && wsURL[:4] == "http" {
		wsURL = "ws" + wsURL[4:]
	}
	return wsURL + "/api/ws/watch"
}

func (c *NucleusClient) APIKey() string {
	return c.apiKey
}
