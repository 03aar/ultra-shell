package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"nucleus-api/models"
	redispkg "nucleus-api/redis"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins in dev
	},
}

var RedisSubscriber *redispkg.Subscriber

func WSStream(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}
	defer conn.Close()

	clientID := uuid.New().String()
	log.Printf("WebSocket client connected: %s", clientID)

	// Send last 10 executions as bootstrap
	executionsMu.RLock()
	bootstrap := make([]models.ExecutionNode, 0)
	startIdx := len(executions) - 10
	if startIdx < 0 {
		startIdx = 0
	}
	for _, exec := range executions[startIdx:] {
		bootstrap = append(bootstrap, exec)
	}
	executionsMu.RUnlock()

	for _, exec := range bootstrap {
		msg := models.WSMessage{
			Type: "execution_complete",
			Data: exec,
		}
		if data, err := json.Marshal(msg); err == nil {
			if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
				return
			}
		}
	}

	// Subscribe to Redis events
	if RedisSubscriber != nil {
		ch := RedisSubscriber.AddListener(clientID)
		defer RedisSubscriber.RemoveListener(clientID)

		// Read pump (handle pings/close)
		done := make(chan struct{})
		go func() {
			defer close(done)
			for {
				_, _, err := conn.ReadMessage()
				if err != nil {
					return
				}
			}
		}()

		// Write pump
		for {
			select {
			case msg, ok := <-ch:
				if !ok {
					return
				}
				data, err := json.Marshal(msg)
				if err != nil {
					continue
				}
				if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
					return
				}
			case <-done:
				return
			}
		}
	} else {
		// No Redis, just keep connection open
		for {
			_, _, err := conn.ReadMessage()
			if err != nil {
				break
			}
		}
	}
}
