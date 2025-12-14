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
	"gorm.io/gorm"
)

func Router(a *gin.Engine, db *gorm.DB, fb *firebase.App) {
	fbService, err := service.NewFirebaseService(fb)

	if err != nil {
		log.Fatal("❌ Could not instantiate firebase service:", err)
	}

	// Init repositories
	userRepo := repository.NewUserRepository(db)
	messageRepo := repository.NewMessageRepository(db)
	serverRepo := repository.NewServerRepository(db)
	appRepo := repository.NewAppRepository(db)

	// Init services
	userService := service.NewUserService(userRepo)
	messageService := service.NewMessageService(messageRepo)
	serverService := service.NewServerService(serverRepo)
	automationService := service.NewServerAutomationService(serverRepo)
	appService := service.NewAppService(appRepo, serverRepo, automationService)

	// Init handlers
	userHandler := handler.NewUserHandler(userService, *fbService)
	serverHandler := handler.NewServerHandler(serverService, userService)
	appHandler := handler.NewAppHandler(appService)
	automationHandler := handler.NewServerAutomationHandler(serverService, automationService)

	// WebSocket hub
	hub := websocket.NewHub()
	go hub.Run()
	wsHandler := handler.NewWsHandler(hub, messageService, userService)

	// Public routes
	a.POST("/user/profile", userHandler.CreateNewAccount)

	// Protected routes
	authMiddleware := middleware.FirebaseAuthMiddleware(fbService)
	userAuth := a.Group("/user", authMiddleware)
	messageAuth := a.Group("/message", authMiddleware)
	serverAuth := a.Group("/servers", authMiddleware)
	appAuth := a.Group("/apps", authMiddleware)

	{
		userAuth.GET("", userHandler.GetAll)
		userAuth.GET("/:id", userHandler.GetByID)
	}
	{
		messageAuth.GET("/ws", wsHandler.Handle)
	}
	{
		serverAuth.GET("", serverHandler.FindAll)
		serverAuth.POST("", serverHandler.Create)
		serverAuth.GET("/:id", serverHandler.GetByID)
		serverAuth.PUT("/:id", serverHandler.Update)
		serverAuth.DELETE("/:id", serverHandler.Delete)

		// Server automation endpoints
		serverAuth.POST("/:id/check-connection", automationHandler.CheckConnection)
		serverAuth.POST("/:id/install-k8s", automationHandler.InstallK8s)
	}
	{
		appAuth.GET("", appHandler.FindAll)
		appAuth.POST("", appHandler.Create)
		appAuth.GET("/:id", appHandler.GetByID)
		appAuth.GET("/server/:serverId", appHandler.GetByServerID)
		appAuth.PUT("/:id", appHandler.Update)
		appAuth.DELETE("/:id", appHandler.Delete)
	}
}
