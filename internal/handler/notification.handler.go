package handler

import (
	contextkey "backend/internal/common/contextKey"
	"backend/internal/response"
	service "backend/internal/service"

	"firebase.google.com/go/auth"
	"github.com/gin-gonic/gin"
)

type NotificationHandler struct {
	service     service.NotificationService
	userService service.UserService
}

func NewNotificationHandler(service service.NotificationService, userService service.UserService) *NotificationHandler {
	return &NotificationHandler{
		service:     service,
		userService: userService,
	}
}

func (h *NotificationHandler) SubscribeHandler(ctx *gin.Context) {
	var sub service.Subscription
	if err := ctx.ShouldBindJSON(&sub); err != nil {
		response.BadRequest(ctx, err.Error())
		return
	}
	auth := ctx.Request.Context().Value(contextkey.UserFirebase).(*auth.Token)

	user, err := h.userService.FindUserByUID(auth.UID)
	if err != nil {
		response.Forbidden(ctx, err.Error())
		return
	}
	err = h.service.SubscribeHandler(&sub, user)
	if err != nil {
		response.InternalError(ctx, err.Error())
		return
	}
	response.Success(ctx, "Subscribed successfully")
}
