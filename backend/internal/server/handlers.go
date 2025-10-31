package server

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/google/uuid"
)

// Configure WebSocket upgrader
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// ✅ Allow requests from your frontend
		return r.Header.Get("Origin") == "http://localhost:5173"
	},
}

// Register all HTTP routes
func RegisterRoutes(r *gin.Engine, hub *Hub) {
	// ✅ Enable simple CORS headers (frontend access)
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "http://localhost:5173")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	// 🎮 WebSocket connection endpoint
	r.GET("/ws", func(c *gin.Context) {
		serveWs(hub, c.Writer, c.Request)
	})

	// 🏆 Leaderboard endpoint — directly uses Mongo from hub.Store
	r.GET("/leaderboard", func(c *gin.Context) {
		results := hub.Store.GetLeaderboard()
		if results == nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load leaderboard"})
			return
		}
		c.JSON(http.StatusOK, results)
	})
}

// WebSocket connection handler
func serveWs(h *Hub, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("❌ WebSocket upgrade error:", err)
		return
	}

	username := r.URL.Query().Get("username")
	if username == "" {
		username = "Guest"
	}

	playerID := uuid.New().String()

	// ✅ Updated field name (Conn instead of WS)
	client := &Connection{
		Conn: conn, // 👈 main fix here
		Send: make(chan []byte, 256),
	}

	player := &Player{
		ID:       playerID,
		Username: username,
		Conn:     client,
	}

	h.register <- player

	// Start read/write pumps
	go client.WritePump()
	go client.ReadPump(h, player)

}
