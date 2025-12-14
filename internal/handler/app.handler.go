package handler

import (
	"backend/internal/model"
	"backend/internal/response"
	"backend/internal/service"
	"strconv"

	"github.com/gin-gonic/gin"
)

type AppHandler struct {
	appSvc service.AppService
}

func NewAppHandler(appSvc service.AppService) *AppHandler {
	return &AppHandler{
		appSvc: appSvc,
	}
}

func (h *AppHandler) FindAll(c *gin.Context) {
	apps, err := h.appSvc.GetAllApps()
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, apps)
}

func (h *AppHandler) Create(c *gin.Context) {
	var app model.App
	if err := c.ShouldBindJSON(&app); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if err := h.appSvc.CreateApp(&app); err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, app)
}

func (h *AppHandler) GetByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "invalid app ID")
		return
	}

	app, err := h.appSvc.GetAppByID(uint(id))
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	if app == nil {
		response.NotFound(c, "app not found")
		return
	}

	response.Success(c, app)
}

func (h *AppHandler) GetByServerID(c *gin.Context) {
	serverIDStr := c.Param("serverId")
	serverID, err := strconv.ParseUint(serverIDStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "invalid server ID")
		return
	}

	apps, err := h.appSvc.GetAppsByServerID(uint(serverID))
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, apps)
}

func (h *AppHandler) Update(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "invalid app ID")
		return
	}

	var app model.App
	if err := c.ShouldBindJSON(&app); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	app.ID = uint(id)

	if err := h.appSvc.UpdateApp(&app); err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, app)
}

func (h *AppHandler) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "invalid app ID")
		return
	}

	if err := h.appSvc.DeleteApp(uint(id)); err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, gin.H{"message": "app deleted successfully"})
}
