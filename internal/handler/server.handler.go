package handler

import (
	contextkey "backend/internal/common/contextKey"
	"backend/internal/model"
	"backend/internal/response"
	"backend/internal/service"
	"strconv"

	"firebase.google.com/go/auth"
	"github.com/gin-gonic/gin"
)

type ServerHandler struct {
	serverSvc service.ServerService
	userSvc   service.UserService
}

func NewServerHandler(serverSvc service.ServerService, userSvc service.UserService) *ServerHandler {
	return &ServerHandler{
		serverSvc: serverSvc,
		userSvc:   userSvc,
	}
}

func (h *ServerHandler) FindAll(c *gin.Context) {
	servers, err := h.serverSvc.GetAllServers()
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, servers)
}

func (h *ServerHandler) Create(c *gin.Context) {
	var server model.Server
	if err := c.ShouldBindJSON(&server); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	// Get user ID from context (set by auth middleware)
	token := c.Request.Context().Value(contextkey.UserFirebase).(*auth.Token)

	user, err := h.userSvc.FindUserByUID(token.UID)

	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	if user == nil {
		response.NotFound(c, "user not found")
		return
	}
	server.CreatedBy = user.ID

	if err := h.serverSvc.CreateServer(&server); err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, server)
}

func (h *ServerHandler) GetByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "invalid server ID")
		return
	}

	server, err := h.serverSvc.GetServerByID(uint(id))
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	if server == nil {
		response.NotFound(c, "server not found")
		return
	}

	response.Success(c, server)
}

func (h *ServerHandler) Update(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "invalid server ID")
		return
	}

	var server model.Server
	if err := c.ShouldBindJSON(&server); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	server.ID = uint(id)

	if err := h.serverSvc.UpdateServer(&server); err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, server)
}

func (h *ServerHandler) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "invalid server ID")
		return
	}

	if err := h.serverSvc.DeleteServer(uint(id)); err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, gin.H{"message": "server deleted successfully"})
}
