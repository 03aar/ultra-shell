package handlers

import (
	"nucleus-api/models"
	"strings"

	"github.com/gin-gonic/gin"
)

type SearchResult struct {
	Executions []models.ExecutionNode `json:"executions"`
	Total      int                    `json:"total"`
	Query      string                 `json:"query"`
}

func SearchExecutions(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		respondSuccess(c, SearchResult{
			Executions: []models.ExecutionNode{},
			Total:      0,
			Query:      "",
		}, "")
		return
	}

	queryLower := strings.ToLower(query)

	executionsMu.RLock()
	defer executionsMu.RUnlock()

	results := make([]models.ExecutionNode, 0)
	for _, exec := range executions {
		if matchesSearch(exec, queryLower) {
			results = append(results, exec)
		}
	}

	// Reverse to get newest first
	for i, j := 0, len(results)-1; i < j; i, j = i+1, j-1 {
		results[i], results[j] = results[j], results[i]
	}

	// Limit to 100
	if len(results) > 100 {
		results = results[:100]
	}

	respondSuccess(c, SearchResult{
		Executions: results,
		Total:      len(results),
		Query:      query,
	}, "")
}

func matchesSearch(exec models.ExecutionNode, query string) bool {
	if strings.Contains(strings.ToLower(exec.Command.Raw), query) {
		return true
	}
	if strings.Contains(strings.ToLower(exec.Stdout), query) {
		return true
	}
	if strings.Contains(strings.ToLower(exec.Stderr), query) {
		return true
	}
	if strings.Contains(strings.ToLower(exec.Command.Binary), query) {
		return true
	}
	for _, f := range exec.FilesMutated {
		if strings.Contains(strings.ToLower(f), query) {
			return true
		}
	}
	return false
}
