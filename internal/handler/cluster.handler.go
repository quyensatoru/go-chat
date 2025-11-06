package handler

import (
	contextkey "backend/internal/common/contextKey"
	"backend/internal/response"
	"backend/internal/service"
	"log"

	"firebase.google.com/go/auth"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
)

type ClusterHandler struct {
	svc         *service.ClusterService
	userService *service.UserService
}

func NewClusterHandler(svc *service.ClusterService, userService *service.UserService) *ClusterHandler {
	return &ClusterHandler{svc: svc, userService: userService}
}

func (h *ClusterHandler) Create(c *gin.Context) {
	auth, ok := c.Request.Context().Value(contextkey.UserFirebase).(*auth.Token)
	if !ok {
		log.Printf("❌ Failed to get auth %v", auth)
		response.Forbidden(c, "Unauthorization")
		return
	}

	email := auth.Claims["email"].(string)

	user, err := h.userService.GetByEmail(c, email)

	if err != nil {
		log.Printf("❌ Not found user %v", err)
		response.BadRequest(c, err)
	}

	var cluster bson.M

	if err := c.ShouldBindJSON(&cluster); err != nil {
		response.BadRequest(c, "invalid request body")
		return
	}

	cluster["created_by"] = user.ID
	cluster["status"] = "pending"

	log.Printf("user %v", cluster)

	resultId, err := h.svc.CreateCluster(c, cluster)

	if err != nil {
		response.InternalError(c, err)
		return
	}

	payload := map[string]interface{}{
		"insertedId": resultId.InsertedID,
	}

	response.Success(c, payload)
}

func (h *ClusterHandler) FindOne(c *gin.Context) {
	var filter bson.M
	if err := c.ShouldBindJSON(&filter); err != nil {
		response.BadRequest(c, "invalid request body")
		return
	}
	cluster, err := h.svc.FindOne(c, filter)
	if err != nil {
		response.InternalError(c, err)
		return
	}
	response.Success(c, cluster)
}

func (h *ClusterHandler) FindAll(c *gin.Context) {
	clusters, err := h.svc.FindAll(c, nil)
	if err != nil {
		response.InternalError(c, err)
		return
	}
	response.Success(c, clusters)
}
