package config

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func ConnectDB() *mongo.Database {
	env := LoadEnv()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(env.DatabaseURI))
	if err != nil {
		log.Fatal("❌ Failed to connect to MongoDB:", err)
	}

	// Kiểm tra kết nối
	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatal("❌ Could not ping MongoDB:", err)
	}

	log.Println("✅ Connected to MongoDB")

	return client.Database("golang_toturial")
}
