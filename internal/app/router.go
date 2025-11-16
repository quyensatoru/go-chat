package app

import (
	"backend/internal/handler"
	"backend/internal/middleware"
	"backend/internal/repository"
	"backend/internal/service"
	"backend/internal/websocket"
	"log"

	firebase "firebase.google.com/go"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

func Router(a *gin.Engine, db *mongo.Database, fb *firebase.App) {
	fbService, err := service.NewFirebaseService(fb)

	if err != nil {
		log.Fatal("❌ Could new instant firebase service:", err)
	}
	// init collection
	collectionUser := db.Collection("users")
	collectionMessage := db.Collection("messages")
	collectionConversation := db.Collection("conversations")
	colectionCluster := db.Collection("clusters")
	collectionDeployment := db.Collection("deployments")
	collectionEviroment := db.Collection("enviroments")

	// init repo
	userRepo := repository.NewUserRepository(collectionUser)
	messageRepo := repository.NewMessageRepository(collectionMessage)
	conversationRepo := repository.NewConversationRepository(collectionConversation)
	clusterRepo := repository.NewClusterRepository(colectionCluster)
	deploymentRepo := repository.NewDeploymentRepository(collectionDeployment)
	enviromentRepo := repository.NewEnviromentRepository(collectionEviroment)

	// init service
	userService := service.NewUserService(userRepo)
	messageService := service.NewMessageService(messageRepo)
	conversationService := service.NewConversationService(conversationRepo)
	clusterService := service.NewClusterService(clusterRepo)
	autoService := service.NewAutomationService(clusterRepo)
	deploymentService := service.NewDeploymentService(deploymentRepo)
	enviromentService := service.NewEnviromentService(enviromentRepo)

	// init handler
	userHandler := handler.NewUserHandler(userService, fbService)
	hub := websocket.NewHub()
	go hub.Run()
	wsHandler := handler.NewWsHandler(hub, messageService, userService, conversationService)
	clusterHandler := handler.NewClusterHandler(
		clusterService,
		userService,
		deploymentService,
		enviromentService,
		autoService,
	)
	deploymentHandler := handler.NewDeploymentHandler(deploymentService, enviromentService, autoService)

	a.POST("/user/profile", userHandler.CreateNewAccount)

	authMiddleware := middleware.FirebaseAuthMiddleware(fbService)
	userAuth := a.Group("/user", authMiddleware)
	messageAuth := a.Group("/message", authMiddleware)
	clusterAuth := a.Group("/clusters", authMiddleware)
	deploymentAuth := a.Group("/deployments", authMiddleware)
	{
		userAuth.GET("", userHandler.GetAll)
	}
	{
		messageAuth.GET("/ws", wsHandler.Handle)
	}
	{
		clusterAuth.GET("", clusterHandler.FindAll)
		clusterAuth.POST("", clusterHandler.Create)
	}
	{
		deploymentAuth.GET("", deploymentHandler.FindAll)
		deploymentAuth.POST("", deploymentHandler.Create)
	}
}
