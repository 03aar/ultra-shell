package handlers

import (
	"net/http"
	"nucleus-api/models"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// In-memory execution store (populated via Redis subscription from nucleus-core)
var (
	executions   []models.ExecutionNode
	executionsMu sync.RWMutex
	edges        []models.Edge
)

func AddExecution(exec models.ExecutionNode) {
	executionsMu.Lock()
	defer executionsMu.Unlock()
	executions = append(executions, exec)
}

func AddEdge(edge models.Edge) {
	executionsMu.Lock()
	defer executionsMu.Unlock()
	edges = append(edges, edge)
}

func GetExecutions(c *gin.Context) {
	executionsMu.RLock()
	defer executionsMu.RUnlock()

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	sessionID := c.Query("session_id")
	riskLevel := c.Query("risk_level")
	category := c.Query("command_category")

	filtered := make([]models.ExecutionNode, 0)
	for _, exec := range executions {
		if sessionID != "" && exec.SessionID != sessionID {
			continue
		}
		if riskLevel != "" {
			hasRisk := false
			for _, flag := range exec.RiskFlags {
				if flag.Level == riskLevel {
					hasRisk = true
					break
				}
			}
			if !hasRisk {
				continue
			}
		}
		if category != "" && exec.Command.Category != category {
			continue
		}
		filtered = append(filtered, exec)
	}

	total := len(filtered)

	// Apply offset and limit
	if offset > len(filtered) {
		filtered = []models.ExecutionNode{}
	} else {
		filtered = filtered[offset:]
	}
	if limit < len(filtered) {
		filtered = filtered[:limit]
	}

	// Reverse to get newest first
	for i, j := 0, len(filtered)-1; i < j; i, j = i+1, j-1 {
		filtered[i], filtered[j] = filtered[j], filtered[i]
	}

	sid := ""
	if len(filtered) > 0 {
		sid = filtered[0].SessionID
	}

	respondSuccess(c, models.ExecutionListResponse{
		Executions: filtered,
		Total:      total,
		SessionID:  sid,
	}, sid)
}

func GetExecution(c *gin.Context) {
	id := c.Param("id")

	executionsMu.RLock()
	defer executionsMu.RUnlock()

	for _, exec := range executions {
		if exec.ID == id {
			respondSuccess(c, exec, exec.SessionID)
			return
		}
	}

	errMsg := "Execution not found"
	c.JSON(http.StatusNotFound, models.APIResponse{
		Data:  nil,
		Meta:  models.APIMeta{Timestamp: time.Now().UTC().Format(time.RFC3339), Version: "1.0"},
		Error: &errMsg,
	})
}

func GetExecutionContext(c *gin.Context) {
	id := c.Param("id")

	executionsMu.RLock()
	defer executionsMu.RUnlock()

	var found *models.ExecutionNode
	var foundIdx int
	for i, exec := range executions {
		if exec.ID == id {
			found = &executions[i]
			foundIdx = i
			break
		}
	}

	if found == nil {
		errMsg := "Execution not found"
		c.JSON(http.StatusNotFound, models.APIResponse{
			Data:  nil,
			Meta:  models.APIMeta{Timestamp: time.Now().UTC().Format(time.RFC3339), Version: "1.0"},
			Error: &errMsg,
		})
		return
	}

	// Get related edges
	relatedEdges := make([]models.Edge, 0)
	for _, edge := range edges {
		if edge.From == id || edge.To == id {
			relatedEdges = append(relatedEdges, edge)
		}
	}

	// Get 5 most recent prior executions
	priorStart := foundIdx - 5
	if priorStart < 0 {
		priorStart = 0
	}
	recentPrior := executions[priorStart:foundIdx]

	resp := models.ExecutionContextResponse{
		Execution:         *found,
		Edges:             relatedEdges,
		RecentPrior:       recentPrior,
		EnvDiffSinceStart: make(map[string]string),
	}

	respondSuccess(c, resp, found.SessionID)
}

func respondSuccess(c *gin.Context, data interface{}, sessionID string) {
	c.JSON(http.StatusOK, models.APIResponse{
		Data: data,
		Meta: models.APIMeta{
			Timestamp: time.Now().UTC().Format(time.RFC3339),
			SessionID: sessionID,
			Version:   "1.0",
		},
		Error: nil,
	})
}
