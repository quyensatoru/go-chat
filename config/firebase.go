package config

import (
	"context"
	"log"
	"os"
	"path/filepath"

	firebase "firebase.google.com/go"
	"google.golang.org/api/option"
)

func ConnectFirebase() *firebase.App {
	ctx := context.Background()

	dir, err := os.Getwd()

	if err != nil {
		log.Fatal("❌ Could not find work directory:", err)
	}

	data, err := os.ReadFile(filepath.Join(dir, "serviceAccountKey.json"))

	if err != nil {
		log.Fatal("❌ Could not find config firebase:", err)
	}

	opt := option.WithCredentialsJSON(data)

	app, err := firebase.NewApp(ctx, nil, opt)

	if err != nil {
		log.Fatal("❌ Could not connect Firebase:", err)
	}

	log.Println("✅ Connected to Firebase")

	return app
}
