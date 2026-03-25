package handlers

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"nucleus-api/models"
	"os"
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
	Source      string `json:"source"`
}

// NaturalTranslate converts natural language to a shell command.
// Tries the configured LLM provider first (via ANTHROPIC_API_KEY, OPENAI_API_KEY,
// or OLLAMA_HOST env vars), then falls back to built-in pattern matching.
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

	// Try LLM translation first
	if result, ok := translateViaLLM(req.Input); ok && result.Command != "" {
		respondSuccess(c, result, "")
		return
	}

	// Fallback to built-in pattern matching
	result := translateBuiltin(req.Input)
	respondSuccess(c, result, "")
}

const llmSystemPrompt = `You are a shell command translator. Convert natural language to a single shell command.
Return ONLY valid JSON: {"command": "the_command", "explanation": "what it does", "risk_level": "low|medium|high"}
No markdown, no code fences, no extra text. Just the JSON object.`

// translateViaLLM calls the configured LLM provider.
// Checks env vars: ANTHROPIC_API_KEY -> OPENAI_API_KEY -> OLLAMA_HOST (in order).
func translateViaLLM(input string) (TranslateResponse, bool) {
	userPrompt := "Translate to a shell command: " + input

	if key := os.Getenv("ANTHROPIC_API_KEY"); key != "" {
		if result, err := callAnthropic(key, userPrompt); err == nil {
			return result, true
		}
	}

	if key := os.Getenv("OPENAI_API_KEY"); key != "" {
		if result, err := callOpenAI(key, userPrompt); err == nil {
			return result, true
		}
	}

	ollamaHost := os.Getenv("OLLAMA_HOST")
	if ollamaHost == "" {
		ollamaHost = "http://localhost:11434"
	}
	if resp, err := http.Get(ollamaHost + "/api/tags"); err == nil {
		resp.Body.Close()
		if result, err := callOllama(ollamaHost, userPrompt); err == nil {
			return result, true
		}
	}

	return TranslateResponse{}, false
}

func callAnthropic(apiKey, userPrompt string) (TranslateResponse, error) {
	body, _ := json.Marshal(map[string]interface{}{
		"model":      "claude-sonnet-4-5-20250514",
		"max_tokens": 256,
		"system":     llmSystemPrompt,
		"messages":   []map[string]string{{"role": "user", "content": userPrompt}},
	})
	req, _ := http.NewRequest("POST", "https://api.anthropic.com/v1/messages", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := (&http.Client{Timeout: 15 * time.Second}).Do(req)
	if err != nil {
		return TranslateResponse{}, err
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)

	var parsed struct {
		Content []struct{ Text string `json:"text"` } `json:"content"`
	}
	json.Unmarshal(respBody, &parsed)
	if len(parsed.Content) > 0 {
		return parseLLMJSON(parsed.Content[0].Text, "llm:anthropic"), nil
	}
	return TranslateResponse{}, nil
}

func callOpenAI(apiKey, userPrompt string) (TranslateResponse, error) {
	body, _ := json.Marshal(map[string]interface{}{
		"model":      "gpt-4o-mini",
		"max_tokens": 256,
		"messages": []map[string]string{
			{"role": "system", "content": llmSystemPrompt},
			{"role": "user", "content": userPrompt},
		},
	})
	req, _ := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := (&http.Client{Timeout: 15 * time.Second}).Do(req)
	if err != nil {
		return TranslateResponse{}, err
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)

	var parsed struct {
		Choices []struct{ Message struct{ Content string `json:"content"` } `json:"message"` } `json:"choices"`
	}
	json.Unmarshal(respBody, &parsed)
	if len(parsed.Choices) > 0 {
		return parseLLMJSON(parsed.Choices[0].Message.Content, "llm:openai"), nil
	}
	return TranslateResponse{}, nil
}

func callOllama(host, userPrompt string) (TranslateResponse, error) {
	model := os.Getenv("OLLAMA_MODEL")
	if model == "" {
		model = "llama3"
	}
	body, _ := json.Marshal(map[string]interface{}{
		"model":  model,
		"stream": false,
		"messages": []map[string]string{
			{"role": "system", "content": llmSystemPrompt},
			{"role": "user", "content": userPrompt},
		},
	})
	req, _ := http.NewRequest("POST", host+"/api/chat", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := (&http.Client{Timeout: 30 * time.Second}).Do(req)
	if err != nil {
		return TranslateResponse{}, err
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)

	var parsed struct {
		Message struct{ Content string `json:"content"` } `json:"message"`
	}
	json.Unmarshal(respBody, &parsed)
	if parsed.Message.Content != "" {
		return parseLLMJSON(parsed.Message.Content, "llm:ollama"), nil
	}
	return TranslateResponse{}, nil
}

func parseLLMJSON(text, source string) TranslateResponse {
	text = strings.TrimSpace(text)
	// Strip markdown code fences
	if strings.HasPrefix(text, "```") {
		lines := strings.Split(text, "\n")
		if len(lines) > 2 {
			text = strings.Join(lines[1:len(lines)-1], "\n")
		}
	}
	var result TranslateResponse
	if err := json.Unmarshal([]byte(text), &result); err != nil {
		result.Command = strings.TrimSpace(text)
		result.Explanation = "Translated via LLM"
		result.RiskLevel = "low"
	}
	result.Source = source
	return result
}

// translateBuiltin uses pattern matching for common queries (no LLM needed).
func translateBuiltin(input string) TranslateResponse {
	q := strings.ToLower(strings.TrimSpace(input))

	translations := []struct {
		patterns    []string
		command     string
		explanation string
		risk        string
	}{
		{[]string{"large file", "big file", "files larger"}, "find . -size +100M -type f", "Find all files larger than 100MB", "low"},
		{[]string{"what changed", "recent change"}, "git diff --stat HEAD~5", "Show changes in the last 5 commits", "low"},
		{[]string{"who is using port", "what's on port"}, "lsof -i :3000 || ss -tlnp | grep 3000", "Show processes on port 3000", "low"},
		{[]string{"listening port", "open port"}, "ss -tlnp", "Show all listening TCP ports", "low"},
		{[]string{"disk space", "disk usage"}, "df -h", "Show disk usage", "low"},
		{[]string{"memory", "ram"}, "free -h", "Show memory usage", "low"},
		{[]string{"cpu", "top process"}, "ps aux --sort=-%cpu | head -20", "Show top processes by CPU", "low"},
		{[]string{"undo", "rollback"}, "nuc rollback --last", "Rollback the most recent command", "medium"},
		{[]string{"git status"}, "git status", "Show git status", "low"},
		{[]string{"git log", "commit history"}, "git log --oneline -20", "Show last 20 commits", "low"},
		{[]string{"docker container"}, "docker ps -a", "List Docker containers", "low"},
		{[]string{"env", "environment variable"}, "env | sort | head -50", "Show environment variables", "low"},
		{[]string{"test", "run test"}, "test -f Makefile && make test || test -f package.json && npm test || echo 'No test runner found'", "Run tests", "low"},
		{[]string{"build"}, "test -f Makefile && make build || test -f package.json && npm run build || echo 'No build tool found'", "Run build", "low"},
		{[]string{"system info"}, "uname -a", "Show system information", "low"},
	}

	for _, t := range translations {
		for _, pattern := range t.patterns {
			if strings.Contains(q, pattern) {
				return TranslateResponse{Command: t.command, Explanation: t.explanation, RiskLevel: t.risk, Source: "builtin"}
			}
		}
	}

	return TranslateResponse{
		Command:     "",
		Explanation: "Could not translate. Set ANTHROPIC_API_KEY, OPENAI_API_KEY, or OLLAMA_HOST for LLM-powered translation.",
		RiskLevel:   "low",
		Source:      "none",
	}
}
