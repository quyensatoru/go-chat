package websocket

import (
	"backend/internal/model"
	"backend/internal/service"
	"context"
	"encoding/json"
	"log"
	"sort"
	"time"

	ws "github.com/gorilla/websocket"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Client struct {
	ConversationId      string
	Conn                *ws.Conn
	Send                chan []byte
	Hub                 *Hub
	MsgService          *service.MessageService
	UserService         *service.UserService
	ConversationService *service.ConversationService
	Closed              bool
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
				// Rời channel cũ nếu đang ở trong
				c.Hub.Unregister <- c
			}
			// --- JOIN channel ---
			targetID, err := primitive.ObjectIDFromHex(data["target_id"].(string))
			if err != nil {
				log.Println("⚠️ Missing target_id in join message")
				continue
			}

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			ids := []primitive.ObjectID{sender.ID, targetID}
			sort.Slice(ids, func(i, j int) bool { return ids[i].Hex() < ids[j].Hex() })

			filter := bson.M{
				"user_ids": ids,
			}

			update := bson.M{
				"$setOnInsert": bson.M{
					"created_at": time.Now().Unix(),
				},
				"$set": bson.M{
					"updated_at": time.Now().Unix(),
				},
			}

			conversation, err := c.ConversationService.FindOneAndUpdate(ctx, filter, update, options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After))

			if err != nil {
				log.Printf("❌ Cannot create or get converation: %v", err)
				return
			}

			log.Printf("conversation: %s", conversation)

			c.ConversationId = conversation["_id"].(primitive.ObjectID).Hex()
			c.Hub.Register <- c
			log.Printf("👤 %s joined channel %s", c.ConversationId, targetID)

		case "message":
			// --- gửi message tới tất cả trong channel ---
			if c.ConversationId == "" {
				log.Println("⚠️ Client chưa join channel, bỏ qua message")
				continue
			}

			content, _ := data["data"].(string)
			broadcast := &Broadcast{
				ConversationId: c.ConversationId,
				Data:           message,
			}
			c.Hub.Broadcast <- broadcast

			// Lưu vào DB (async)
			go func() {
				ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
				defer cancel()

				user, err := c.UserService.GetByID(ctx, data["target_id"].(string))
				if err != nil {
					log.Printf("❌ Get recipient failed: %v", err)
					return
				}
				payload := &model.Message{
					Content:     content,
					RecepientID: user.ID,
					SenderID:    sender.ID,
				}
				if _, err := c.MsgService.Create(ctx, payload); err != nil {
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
