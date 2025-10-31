package main

import (
	"log"
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/samariium/Backend_Assignment/backend/internal/server"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("⚠️ No .env file found, using defaults")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "9090"
	}

	frontendOrigin := os.Getenv("FRONTEND_ORIGIN")
	if frontendOrigin == "" {
		frontendOrigin = "http://localhost:5173"
	}

	r := gin.Default()

	// ✅ Enable CORS for frontend
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{frontendOrigin},
		AllowMethods:     []string{"GET", "POST", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// ✅ Create and run the main Hub
	hub := server.NewHub()
	go hub.Run()

	// ✅ Register WebSocket + Leaderboard routes
	server.RegisterRoutes(r, hub)

	// ✅ Start backend server
	log.Printf("✅ Backend running on :%s (CORS allowed: %s)\n", port, frontendOrigin)
	if err := r.Run(":" + port); err != nil {
		log.Fatal("❌ Failed to start server:", err)
	}
}
