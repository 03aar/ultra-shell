package redis

import (
	"context"
	"encoding/json"
	"log"
	"nucleus-api/models"
	"sync"

	goredis "github.com/redis/go-redis/v9"
)

type Subscriber struct {
	client     *goredis.Client
	ctx        context.Context
	cancel     context.CancelFunc
	listeners  map[string][]chan models.WSMessage
	mu         sync.RWMutex
	connected  bool
}

func NewSubscriber(redisURL string) *Subscriber {
	ctx, cancel := context.WithCancel(context.Background())

	opt, err := goredis.ParseURL(redisURL)
	if err != nil {
		log.Printf("Failed to parse Redis URL: %v, using defaults", err)
		opt = &goredis.Options{
			Addr: "redis:6379",
		}
	}

	client := goredis.NewClient(opt)

	// Test connection
	connected := false
	if err := client.Ping(ctx).Err(); err != nil {
		log.Printf("Redis not available: %v (will use in-memory fallback)", err)
	} else {
		connected = true
		log.Println("Connected to Redis")
	}

	return &Subscriber{
		client:    client,
		ctx:       ctx,
		cancel:    cancel,
		listeners: make(map[string][]chan models.WSMessage),
		connected: connected,
	}
}

func (s *Subscriber) Subscribe(channels ...string) {
	if !s.connected {
		log.Println("Redis not connected, skipping subscription")
		return
	}

	pubsub := s.client.Subscribe(s.ctx, channels...)

	go func() {
		defer pubsub.Close()
		ch := pubsub.Channel()

		for msg := range ch {
			s.handleMessage(msg.Channel, msg.Payload)
		}
	}()

	log.Printf("Subscribed to Redis channels: %v", channels)
}

func (s *Subscriber) handleMessage(channel, payload string) {
	var msgType string
	switch channel {
	case "nucleus:executions":
		msgType = "execution_complete"
	case "nucleus:rollbacks":
		msgType = "rollback_complete"
	case "nucleus:sessions":
		msgType = "session_started"
	default:
		msgType = "unknown"
	}

	var data interface{}
	if err := json.Unmarshal([]byte(payload), &data); err != nil {
		data = payload
	}

	wsMsg := models.WSMessage{
		Type: msgType,
		Data: data,
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	// Broadcast to all listeners
	for _, listeners := range s.listeners {
		for _, ch := range listeners {
			select {
			case ch <- wsMsg:
			default:
				// Channel full, skip
			}
		}
	}
}

func (s *Subscriber) AddListener(id string) chan models.WSMessage {
	s.mu.Lock()
	defer s.mu.Unlock()

	ch := make(chan models.WSMessage, 100)
	s.listeners[id] = append(s.listeners[id], ch)
	return ch
}

func (s *Subscriber) RemoveListener(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if channels, ok := s.listeners[id]; ok {
		for _, ch := range channels {
			close(ch)
		}
		delete(s.listeners, id)
	}
}

// Publish sends a message directly (for in-memory fallback when no Redis)
func (s *Subscriber) Publish(msgType string, data interface{}) {
	wsMsg := models.WSMessage{
		Type: msgType,
		Data: data,
	}

	if s.connected {
		jsonData, err := json.Marshal(data)
		if err == nil {
			channel := "nucleus:executions"
			if msgType == "rollback_complete" {
				channel = "nucleus:rollbacks"
			} else if msgType == "session_started" {
				channel = "nucleus:sessions"
			}
			s.client.Publish(s.ctx, channel, string(jsonData))
		}
	}

	// Also broadcast directly to local listeners
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, listeners := range s.listeners {
		for _, ch := range listeners {
			select {
			case ch <- wsMsg:
			default:
			}
		}
	}
}

func (s *Subscriber) Close() {
	s.cancel()
	s.client.Close()
}
