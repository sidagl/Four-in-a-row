package server

import (
	"encoding/json"
	"log"
	"time"

	"github.com/gorilla/websocket"
)

// ---------------------------------------------
// Represents a single WebSocket connection for a player
// ---------------------------------------------
type Connection struct {
	Conn *websocket.Conn
	Send chan []byte
}

// ---------------------------------------------
// Constants for connection management
// ---------------------------------------------
const (
	writeWait  = 10 * time.Second    // Max wait before writing
	pongWait   = 60 * time.Second    // How long to wait for the next pong
	pingPeriod = (pongWait * 9) / 10 // Send ping slightly before pong timeout
)

// ---------------------------------------------
// WritePump: sends queued messages and periodic heartbeats
// ---------------------------------------------
func (c *Connection) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case msg, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// Channel closed — close WS
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			// Send message to client
			if err := c.Conn.WriteMessage(websocket.TextMessage, msg); err != nil {
				log.Println("⚠️ Write error:", err)
				return
			}

		case <-ticker.C:
			// Send periodic JSON heartbeat instead of WS Ping
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			heartbeat := map[string]string{"type": "ping"}
			if err := c.Conn.WriteJSON(heartbeat); err != nil {
				log.Println("⚠️ Heartbeat failed:", err)
				return
			}
		}
	}
}

// ---------------------------------------------
// ReadPump: listens for incoming messages & handles heartbeats
// ---------------------------------------------
func (c *Connection) ReadPump(h *Hub, player *Player) {
	defer func() {
		h.unregister <- player
		c.Conn.Close()
	}()

	c.Conn.SetReadLimit(512)
	c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, msg, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("⚠️ Unexpected WS close for %s: %v", player.Username, err)
			} else {
				log.Printf("ℹ️ Connection closed gracefully for %s: %v", player.Username, err)
			}
			break
		}

		// 🧠 Check if it's a heartbeat pong
		var data map[string]interface{}
		if err := json.Unmarshal(msg, &data); err == nil {
			if data["type"] == "pong" {
				log.Printf("💓 Pong received from %s", player.Username)
				// Refresh read deadline to keep connection alive
				c.Conn.SetReadDeadline(time.Now().Add(pongWait))
				continue
			}
		}

		// 🕹 Forward to the player's current game room
		if player.Room != nil {
			player.Room.HandleIncomingMessage(player, msg)
		}
	}
}
