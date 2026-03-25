package handlers

import (
	"net/http"
	"nucleus-api/models"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

var (
	sessions   []models.Session
	sessionsMu sync.RWMutex
)

func init() {
	sessions = make([]models.Session, 0)
}

func AddSession(session models.Session) {
	sessionsMu.Lock()
	defer sessionsMu.Unlock()
	sessions = append(sessions, session)
}

func GetSessions(c *gin.Context) {
	sessionsMu.RLock()
	defer sessionsMu.RUnlock()

	// Enrich sessions with execution counts
	executionsMu.RLock()
	enriched := make([]models.Session, len(sessions))
	copy(enriched, sessions)
	for i := range enriched {
		execCount := 0
		riskCount := 0
		for _, exec := range executions {
			if exec.SessionID == enriched[i].ID {
				execCount++
				if len(exec.RiskFlags) > 0 {
					riskCount++
				}
			}
		}
		enriched[i].ExecutionCount = execCount
		enriched[i].RiskWarningCount = riskCount
	}
	executionsMu.RUnlock()

	respondSuccess(c, enriched, "")
}

func CreateSession(c *gin.Context) {
	var req models.CreateSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errMsg := "Invalid request: name is required"
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Data:  nil,
			Meta:  models.APIMeta{Timestamp: time.Now().UTC().Format(time.RFC3339), Version: "1.0"},
			Error: &errMsg,
		})
		return
	}

	session := models.Session{
		ID:        uuid.New().String(),
		Name:      req.Name,
		StartTime: time.Now().UTC(),
		CreatedAt: time.Now().UTC(),
	}

	sessionsMu.Lock()
	sessions = append(sessions, session)
	sessionsMu.Unlock()

	respondSuccess(c, session, session.ID)
}

func GetSessionReplay(c *gin.Context) {
	sessionID := c.Param("id")

	sessionsMu.RLock()
	var found *models.Session
	for i, s := range sessions {
		if s.ID == sessionID {
			found = &sessions[i]
			break
		}
	}
	sessionsMu.RUnlock()

	if found == nil {
		errMsg := "Session not found"
		c.JSON(http.StatusNotFound, models.APIResponse{
			Data:  nil,
			Meta:  models.APIMeta{Timestamp: time.Now().UTC().Format(time.RFC3339), Version: "1.0"},
			Error: &errMsg,
		})
		return
	}

	executionsMu.RLock()
	sessionExecs := make([]models.ExecutionNode, 0)
	for _, exec := range executions {
		if exec.SessionID == sessionID {
			sessionExecs = append(sessionExecs, exec)
		}
	}
	executionsMu.RUnlock()

	respondSuccess(c, models.SessionReplayResponse{
		Session:    *found,
		Executions: sessionExecs,
	}, sessionID)
}
