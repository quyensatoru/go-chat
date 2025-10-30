package main

import (
	"backend/config"
	"backend/internal/app"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	db := config.ConnectDB()
	fb := config.ConnectFirebase()

	g := gin.Default()

	// Configure CORS to allow the frontend to send Authorization header and credentials
	g.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000", "http://localhost:3001"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	//initialize routes
	app.Router(g, db, fb)

	g.Run(":3000")
}
