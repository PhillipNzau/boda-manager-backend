package main

import (
	"log"
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"github.com/PhillipNzau/boda-manager-backend/config"
	"github.com/PhillipNzau/boda-manager-backend/routes"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("no .env file loaded, reading environment variables")
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("config load error: %v", err)
	}

	log.Println("✅ Connected to MongoDB")

	config.EnsureAllIndexes(cfg.MongoClient, cfg.DBName)

	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins: []string{
			"https://sub-safe-two.vercel.app",
			"https://www.subsafe.co.ke",
			"http://localhost:4200",
		},
		AllowMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders: []string{
			"Origin",
			"Content-Type",
			"Authorization",
			"If-None-Match",
			"If-Modified-Since",
		},
		ExposeHeaders: []string{
			"ETag",
			"Last-Modified",
			"Content-Length",
		},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	routes.SetupRoutes(r, cfg)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("🚀 Listening on :%s\n", port)
	log.Fatal(r.Run(":" + port))
}
