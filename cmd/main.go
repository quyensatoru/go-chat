package main

import (
	"backend/config"
	"backend/internal/app"
	"fmt"
	"time"

	"github.com/SherClockHolmes/webpush-go"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	db := config.ConnectDB()
	fb := config.ConnectFirebase()
	redis := config.ConnectRedis()

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
	app.Router(g, db, fb, redis)

	privateKey, publicKey, err := webpush.GenerateVAPIDKeys()
	if err != nil {
		panic(err)
	}

	fmt.Println("Public Key:", publicKey)
	fmt.Println("Private Key:", privateKey)

	g.Run(":3000")
}
