package websocket

import (
	"backend/internal/model"
	"backend/internal/service"
	"encoding/json"
	"fmt"
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
	NotifyService  service.NotificationService
	RedisService   service.RedisService
	Closed         bool
}

type WsMessage struct {
	Type     string `json:"type"`
	TargetID int64  `json:"target_id"`
	Content  string `json:"content"`
}

func buildRoomID(user1, user2 int64) string {
	if user1 < user2 {
		return fmt.Sprintf("dm_%d_%d", user1, user2)
	}
	return fmt.Sprintf("dm_%d_%d", user2, user1)
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

		var data WsMessage
		if err := json.Unmarshal(message, &data); err != nil {
			log.Println("⚠️ Invalid JSON:", err)
			continue
		}

		switch data.Type {
		case "join":
			log.Printf("%+v", data)
			if c.ConversationId != "" {
				// Leave old channel if already in one
				c.Hub.Unregister <- c
			}
			// --- JOIN channel ---
			if data.TargetID == 0 {
				log.Printf("⚠️ Missing target_id in join message %v", data.TargetID)
				continue
			}

			// Create conversation ID from sender and target
			channelID := buildRoomID(int64(sender.ID), data.TargetID)
			c.ConversationId = channelID
			c.Hub.Register <- c
			log.Printf("👤 User %d joined channel %d", sender.ID, data.TargetID)

		case "message":
			// --- send message to all in channel ---
			if c.ConversationId == "" {
				log.Println("⚠️ Client hasn't joined channel, skipping message")
				continue
			}

			content := data.Content
			broadcast := &Broadcast{
				ConversationId: c.ConversationId,
				Content:        message,
			}
			c.Hub.Broadcast <- broadcast

			// Save to DB (async)
			go func() {
				targetID := uint(data.TargetID)

				payload := &model.Message{
					Content:     content,
					RecipientID: targetID,
					SenderID:    sender.ID,
					ChannelID:   c.ConversationId,
				}
				if err := c.MsgService.CreateMessage(payload); err != nil {
					log.Printf("⚠️ Error saving message: %v", err)
				}
				// Cache message in Redis
				c.RedisService.CacheMessage(c.ConversationId, model.MessageView{
					ID:        payload.ID,
					Content:   payload.Content,
					TargetID:  payload.RecipientID,
					SenderID:  payload.SenderID,
					CreatedAt: payload.CreatedAt,
				})
				// Send notification to recipient
				c.NotifyService.SendNotification(service.NotificationPayload{
					Title: sender.Username,
					Body:  content,
				}, &model.User{ID: targetID})

			}()

		default:
			log.Println("⚠️ Unknown message type:", data.Type)
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
