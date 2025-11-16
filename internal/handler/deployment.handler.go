package handler

import (
	"backend/internal/response"
	"backend/internal/service"
	"errors"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type DeploymentHandler struct {
	svcDeployment *service.DeploymentService
	svcEnv        *service.EnviromentService
	svcAuto       *service.AutomationService
}

func NewDeploymentHandler(svcDeployment *service.DeploymentService, svcEnv *service.EnviromentService, svcAuto *service.AutomationService) *DeploymentHandler {
	return &DeploymentHandler{
		svcDeployment: svcDeployment,
		svcEnv:        svcEnv,
		svcAuto:       svcAuto,
	}
}

func (h *DeploymentHandler) FindAll(ctx *gin.Context) {
	deployments := h.svcDeployment.FindAll(ctx, map[string]interface{}{})
	response.Created(ctx, deployments)
}

func (h *DeploymentHandler) Create(ctx *gin.Context) {
	var payload struct {
		Data []map[string]interface{} `json:"data"`
	}

	if err := ctx.ShouldBindJSON(&payload); err != nil {
		response.BadRequest(ctx, gin.H{"error": "invalid request body"})
		return
	}
	var deployments []interface{}

	for _, deploy := range payload.Data {
		deployments = append(deployments, deploy)
	}

	results, err := h.svcDeployment.CreateMany(ctx, deployments)
	if err != nil {
		response.InternalError(ctx, err)
		return
	}

	for index, id := range results.InsertedIDs {
		if index >= len(payload.Data) {
			response.InternalError(ctx, errors.New("mismatched inserted ids and payload data"))
			return
		}
		deployment := payload.Data[index]

		var oid primitive.ObjectID
		switch v := id.(type) {
		case primitive.ObjectID:
			oid = v
		case string:
			parsed, perr := primitive.ObjectIDFromHex(v)
			if perr != nil {
				response.InternalError(ctx, perr)
				return
			}
			oid = parsed
		default:
			response.InternalError(ctx, errors.New("unsupported inserted ID type"))
			return
		}

		if _, err := h.svcEnv.Create(ctx, map[string]interface{}{
			"dataset":       deployment["enviroment"],
			"deployment_id": oid,
			"created_at":    time.Now(),
			"updated_at":    time.Now(),
		}); err != nil {
			response.InternalError(ctx, err)
			return
		}
	}

	ctx.JSON(201, gin.H{"inserted_count": len(results.InsertedIDs)})
}
