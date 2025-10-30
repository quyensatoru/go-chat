package websocket

import (
	"log"
	"sync"
)

type Broadcast struct {
	ConversationId string `json:"conversation_id"`
	Data           []byte `json:"data"`
}

type Hub struct {
	Channels   map[string]map[*Client]bool
	Broadcast  chan *Broadcast
	Register   chan *Client
	Unregister chan *Client
	Mu         sync.Mutex
}

func NewHub() *Hub {
	return &Hub{
		Channels:   make(map[string]map[*Client]bool),
		Broadcast:  make(chan *Broadcast),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.Mu.Lock()
			if _, ok := h.Channels[client.ConversationId]; !ok {
				h.Channels[client.ConversationId] = make(map[*Client]bool)
			}
			h.Channels[client.ConversationId][client] = true
			h.Mu.Unlock()

			log.Printf("✅ Joined channel %s (%p)", client.ConversationId, client)
			for channelID, clients := range h.Channels {
				log.Printf("  📦 Channel: %s (%d clients)", channelID, len(clients))
			}

		case client := <-h.Unregister:
			h.Mu.Lock()
			if clients, ok := h.Channels[client.ConversationId]; ok {
				delete(clients, client)
				// if !client.Closed {
				// 	close(client.Send)
				// 	client.Closed = true
				// }
				if len(clients) == 0 {
					delete(h.Channels, client.ConversationId)
				}
			}
			h.Mu.Unlock()

		case message := <-h.Broadcast:
			h.Mu.Lock()
			if clients, ok := h.Channels[message.ConversationId]; ok {
				for client := range clients {
					select {
					case client.Send <- message.Data:
					default:
						// if !client.Closed {
						// 	close(client.Send)
						// 	client.Closed = true
						// }
						delete(clients, client)
					}
				}
			}
			h.Mu.Unlock()
		}
	}
}
