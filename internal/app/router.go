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

	// init repo
	userRepo := repository.NewUserRepository(collectionUser)
	messageRepo := repository.NewMessageRepository(collectionMessage)
	conversationRepo := repository.NewConversationRepository(collectionConversation)
	clusterRepo := repository.NewClusterRepository(colectionCluster)

	// init service
	userService := service.NewUserService(userRepo)
	messageService := service.NewMessageService(messageRepo)
	conversationService := service.NewConversationService(conversationRepo)
	clusterService := service.NewClusterService(clusterRepo)

	// init handler
	userHandler := handler.NewUserHandler(userService, fbService)
	hub := websocket.NewHub()
	go hub.Run()
	wsHandler := handler.NewWsHandler(hub, messageService, userService, conversationService)
	clusterHandler := handler.NewClusterHandler(clusterService, userService)

	a.POST("/user/profile", userHandler.CreateNewAccount)

	authMiddleware := middleware.FirebaseAuthMiddleware(fbService)
	userAuth := a.Group("/user", authMiddleware)
	messageAuth := a.Group("/message", authMiddleware)
	clusterAuth := a.Group("/clusters", authMiddleware)
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
}
