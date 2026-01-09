package handler

import (
	"backend/internal/response"
	"backend/internal/service"

	"github.com/gin-gonic/gin"
)

type Message struct {
	service      service.MessageService
	redisService service.RedisService
}

func NewMessageHandler(service service.MessageService, redisService service.RedisService) *Message {
	return &Message{
		service:      service,
		redisService: redisService,
	}
}

func (h *Message) GetMessagesByChannel(ctx *gin.Context) {
	//get message by channel lazy load by offset and limit

	type GetMessagesRequest struct {
		ChannelID string `form:"channel_id" binding:"required"`
		Offset    int    `form:"offset"`
		Limit     int    `form:"limit"`
	}
	var body GetMessagesRequest
	if err := ctx.ShouldBindQuery(&body); err != nil {
		response.BadRequest(ctx, err.Error())
		return
	}

	messages, err := h.redisService.GetMessage(body.ChannelID, body.Limit, body.Offset)

	if len(messages) > 0 {
		response.Success(ctx, messages)
		return
	}

	messages, err = h.service.GetMessagesByChannel(body.ChannelID, body.Limit, body.Offset)
	if err != nil {
		response.InternalError(ctx, err)
		return
	}

	for _, message := range messages {
		h.redisService.CacheMessage(body.ChannelID, message)
	}

	response.Success(ctx, messages)
}
