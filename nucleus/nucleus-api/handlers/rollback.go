package handlers

import (
	"net/http"
	"nucleus-api/models"
	"time"

	"github.com/gin-gonic/gin"
)

// RollbackHandler stores reference to Redis subscriber for broadcasting
var RollbackBroadcast func(msgType string, data interface{})

func PostRollback(c *gin.Context) {
	executionID := c.Param("execution_id")

	executionsMu.RLock()
	var found *models.ExecutionNode
	for i, exec := range executions {
		if exec.ID == executionID {
			found = &executions[i]
			break
		}
	}
	executionsMu.RUnlock()

	if found == nil {
		errMsg := "Execution not found"
		c.JSON(http.StatusNotFound, models.APIResponse{
			Data:  nil,
			Meta:  models.APIMeta{Timestamp: time.Now().UTC().Format(time.RFC3339), Version: "1.0"},
			Error: &errMsg,
		})
		return
	}

	if !found.RollbackAvailable {
		errMsg := "Rollback not available for this execution"
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Data:  nil,
			Meta:  models.APIMeta{Timestamp: time.Now().UTC().Format(time.RFC3339), Version: "1.0"},
			Error: &errMsg,
		})
		return
	}

	// Mark as rolled back
	executionsMu.Lock()
	for i, exec := range executions {
		if exec.ID == executionID {
			executions[i].RolledBack = true
			executions[i].RollbackAvailable = false
			break
		}
	}
	executionsMu.Unlock()

	result := models.RollbackResponse{
		Success:       true,
		FilesRestored: found.FilesMutated,
		EnvRestored:   found.EnvChanges,
		Message:       "Rollback successful",
	}

	// Broadcast rollback event
	if RollbackBroadcast != nil {
		RollbackBroadcast("rollback_complete", map[string]interface{}{
			"execution_id":   executionID,
			"files_restored": result.FilesRestored,
			"timestamp":      time.Now().UTC().Format(time.RFC3339),
		})
	}

	respondSuccess(c, result, found.SessionID)
}
