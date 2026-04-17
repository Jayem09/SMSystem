package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"smsystem-backend/internal/services"

	"github.com/gin-gonic/gin"
)

type EventHandler struct{}

func NewEventHandler() *EventHandler {
	return &EventHandler{}
}

// Stream SSE events to the browser.
// GET /api/events
// branch_id query param filters events (0 = all branches).
func (h *EventHandler) Stream(c *gin.Context) {
	// Get branch filter (0 = all branches)
	var branchID uint
	if bid := c.Query("branch_id"); bid != "" {
		if parsed, err := strconv.ParseUint(bid, 10, 64); err == nil {
			branchID = uint(parsed)
		}
	}

	// Register client with broadcaster
	broadcaster := services.GetBroadcaster()
	if broadcaster == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Event service unavailable"})
		return
	}
	client := broadcaster.Register(branchID)
	defer broadcaster.Unregister(client)

	// Set SSE headers
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("X-Accel-Buffering", "no")

	// Send initial ping
	c.SSEvent("ping", map[string]interface{}{
		"client_id": client.ID,
		"connected": true,
		"ts":        time.Now().Unix(),
	})
	c.Writer.Flush()

	// Heartbeat ticker (send ping every 25s to keep connection alive)
	heartbeat := time.NewTicker(25 * time.Second)
	defer heartbeat.Stop()

	// Client disconnect channel
	clientGone := c.Request.Context().Done()

	for {
		select {
		case <-clientGone:
			return
		case <-heartbeat.C:
			c.SSEvent("ping", map[string]interface{}{"ts": time.Now().Unix()})
			c.Writer.Flush()
		case event, ok := <-client.Channel:
			if !ok {
				return
			}
			data, err := json.Marshal(event)
			if err != nil {
				continue
			}
			c.Writer.Write([]byte("data: "))
			c.Writer.Write(data)
			c.Writer.Write([]byte("\n\n"))
			c.Writer.Flush()
		}
	}
}
