package handler

import (
	"backend/internal/response"
	"backend/internal/service"
	"strconv"

	"github.com/gin-gonic/gin"
)

type ServerAutomationHandler struct {
	serverSvc     service.ServerService
	automationSvc service.ServerAutomationService
}

func NewServerAutomationHandler(serverSvc service.ServerService, automationSvc service.ServerAutomationService) *ServerAutomationHandler {
	return &ServerAutomationHandler{
		serverSvc:     serverSvc,
		automationSvc: automationSvc,
	}
}

// CheckConnection tests SSH connection to a server
func (h *ServerAutomationHandler) CheckConnection(c *gin.Context) {
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

	// Test connection
	err = h.automationSvc.CheckConnection(server)
	if err != nil {
		response.InternalError(c, gin.H{
			"message": "connection failed",
			"error":   err.Error(),
		})
		return
	}

	// Update server status to active
	server.Status = "active"
	err = h.serverSvc.UpdateServer(server)
	if err != nil {
		response.InternalError(c, "failed to update server status")
		return
	}

	response.Success(c, gin.H{
		"message": "connection successful",
		"status":  "active",
	})
}

// InstallK8s installs K3s, Helm, and ArgoCD on the server
func (h *ServerAutomationHandler) InstallK8s(c *gin.Context) {
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

	// Parse request body
	var req struct {
		GitBranch           string `json:"git_branch"`
		ArgoCDAdminPassword string `json:"argocd_admin_password"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	// Run installation in background to avoid timeout
	go func() {
		err := h.automationSvc.InstallK8s(server, req.GitBranch, req.ArgoCDAdminPassword)
		if err != nil {
			// Log error
			println("K8s installation failed:", err.Error())
			server.Status = "installation_failed"
		} else {
			server.Status = "k8s_ready"
		}
		h.serverSvc.UpdateServer(server)
	}()

	// Update status to installing
	server.Status = "installing_k8s"
	err = h.serverSvc.UpdateServer(server)
	if err != nil {
		response.InternalError(c, "failed to update server status")
		return
	}

	response.Success(c, gin.H{
		"message": "K8s installation started",
		"status":  "installing_k8s",
	})
}
