package websocket

import (
	"backend/internal/model"
	"backend/internal/service"
	"encoding/json"
	"log"
	"time"

	ws "github.com/gorilla/websocket"
)

type Client struct {
	ConversationId string
	Conn           *ws.Conn
	Send           chan []byte
	Hub            *Hub
	MsgService     service.MessageService
	UserService    service.UserService
	Closed         bool
}

// ------------------------- ReadPump -------------------------
func (c *Client) ReadPump(sender *model.User) {
	defer func() {
		c.Hub.Unregister <- c
		if !c.Closed {
			close(c.Send)
			c.Closed = true
		}
		c.Conn.Close()
	}()

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			break
		}

		var data map[string]interface{}
		if err := json.Unmarshal(message, &data); err != nil {
			log.Println("⚠️ Invalid JSON:", err)
			continue
		}

		switch data["type"] {
		case "join":
			if c.ConversationId != "" {
				// Leave old channel if already in one
				c.Hub.Unregister <- c
			}
			// --- JOIN channel ---
			targetIDStr, ok := data["target_id"].(string)
			if !ok {
				log.Println("⚠️ Missing target_id in join message")
				continue
			}

			// Create conversation ID from sender and target
			c.ConversationId = targetIDStr
			c.Hub.Register <- c
			log.Printf("👤 User %d joined channel %s", sender.ID, targetIDStr)

		case "message":
			// --- send message to all in channel ---
			if c.ConversationId == "" {
				log.Println("⚠️ Client hasn't joined channel, skipping message")
				continue
			}

			content, _ := data["data"].(string)
			broadcast := &Broadcast{
				ConversationId: c.ConversationId,
				Data:           message,
			}
			c.Hub.Broadcast <- broadcast

			// Save to DB (async)
			go func() {
				targetIDFloat, ok := data["target_id"].(float64)
				if !ok {
					log.Printf("❌ Invalid target_id type")
					return
				}
				targetID := uint(targetIDFloat)

				payload := &model.Message{
					Content:     content,
					RecipientID: targetID,
					SenderID:    sender.ID,
					TaskID:      c.ConversationId,
				}
				if err := c.MsgService.CreateMessage(payload); err != nil {
					log.Printf("⚠️ Error saving message: %v", err)
				}
			}()

		default:
			log.Println("⚠️ Unknown message type:", data["type"])
		}
	}
}

// ------------------------- WritePump -------------------------
func (c *Client) WritePump() {
	defer c.Conn.Close()

	const (
		writeWait  = 10 * time.Second
		pongWait   = 60 * time.Second
		pingPeriod = (pongWait * 9) / 10
	)

	c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	ticker := time.NewTicker(pingPeriod)
	defer ticker.Stop()

	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.Conn.WriteMessage(ws.CloseMessage, []byte{})
				return
			}
			if err := c.Conn.WriteMessage(ws.TextMessage, message); err != nil {
				return
			}

		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Conn.WriteMessage(ws.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
