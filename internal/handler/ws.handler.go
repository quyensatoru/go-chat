package handler

import (
	contextkey "backend/internal/common/contextKey"
	"backend/internal/response"
	"backend/internal/service"
	wsService "backend/internal/websocket"
	"log"
	"net/http"

	"firebase.google.com/go/auth"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin:     func(r *http.Request) bool { return true },
	WriteBufferSize: 1024,
	ReadBufferSize:  1024,
}

type WsHandler struct {
	hub                 *wsService.Hub
	msgService          *service.MessageService
	userService         *service.UserService
	conversationService *service.ConversationService
}

func NewWsHandler(h *wsService.Hub, msgService *service.MessageService, userService *service.UserService, converationService *service.ConversationService) *WsHandler {
	return &WsHandler{
		hub:                 h,
		msgService:          msgService,
		userService:         userService,
		conversationService: converationService,
	}
}

func (h *WsHandler) Handle(ctx *gin.Context) {
	auth, ok := ctx.Request.Context().Value(contextkey.UserFirebase).(*auth.Token)
	if !ok {
		log.Printf("❌ Failed to get auth %v", auth)
		response.Forbidden(ctx, "Unauthorization")
		return
	}

	senderEmail := auth.Claims["email"].(string)
	sender, err := h.userService.GetByEmail(ctx, senderEmail)
	if err != nil {
		log.Printf("❌ Failed to get sender: %v\n", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "user not found"})
		return
	}

	conn, err := upgrader.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	client := &wsService.Client{
		Hub:                 h.hub,
		Conn:                conn,
		Send:                make(chan []byte, 256),
		MsgService:          h.msgService,
		UserService:         h.userService,
		ConversationService: h.conversationService,
	}

	go client.WritePump()
	go client.ReadPump(sender)
}
