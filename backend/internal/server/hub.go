package server

import (
	"encoding/json"
	"log"

	"github.com/samariium/Backend_Assignment/backend/internal/storage"
)

// Hub manages all connected players and game rooms
type Hub struct {
	players    map[string]*Player
	register   chan *Player
	unregister chan *Player
	moves      chan MoveMessage
	rooms      map[string]*Room
	Store      *storage.Storage // ✅ MongoDB leaderboard integration
}

// MoveMessage represents a move from a player
type MoveMessage struct {
	PlayerID string `json:"playerId"`
	Column   int    `json:"column"`
}

// NewHub initializes and returns a new Hub
func NewHub() *Hub {
	store := storage.NewStorage() // ✅ Centralized MongoDB connection
	return &Hub{
		players:    make(map[string]*Player),
		register:   make(chan *Player),
		unregister: make(chan *Player),
		moves:      make(chan MoveMessage),
		rooms:      make(map[string]*Room),
		Store:      store,
	}
}

// Run continuously listens for player and move events
func (h *Hub) Run() {
	for {
		select {
		case player := <-h.register:
			h.handleRegister(player)
		case player := <-h.unregister:
			h.handleUnregister(player)
		case move := <-h.moves:
			h.handleMove(move)
		}
	}
}

// 🧍 When a player joins
func (h *Hub) handleRegister(p *Player) {
	log.Printf("🧍 Player joined: %s", p.Username)
	h.players[p.Username] = p

	var opponent *Player
	for name, pl := range h.players {
		if name != p.Username && pl.Room == nil {
			opponent = pl
			break
		}
	}

	if opponent != nil {
		room := NewRoom(p, opponent)
		room.Hub = h
		h.rooms[room.ID] = room
		go room.StartGame()
	}
}


// 👋 When a player disconnects
func (h *Hub) handleUnregister(p *Player) {
	if _, ok := h.players[p.ID]; ok {
		delete(h.players, p.ID)
		close(p.Conn.Send)
		log.Printf("👋 Player left: %s", p.Username)
	}
}

// 🎯 Handle a move event from frontend
func (h *Hub) handleMove(move MoveMessage) {
	player := h.findPlayerByID(move.PlayerID)
	if player == nil || player.Room == nil {
		log.Printf("⚠️ Move ignored — player not found or not in room (id: %s)", move.PlayerID)
		return
	}
	player.Room.HandleMove(player, move.Column)
}

// 🔍 Helper: find player by unique ID
func (h *Hub) findPlayerByID(id string) *Player {
	for _, p := range h.players {
		if p.ID == id {
			return p
		}
	}
	return nil
}

// 📢 Broadcast current game state to both players
func (h *Hub) broadcastGameState(room *Room) {
	msg, _ := json.Marshal(map[string]interface{}{
		"type":   "state",
		"gameId": room.ID,
		"board":  room.G.Board,
		"turn":   room.G.Turn,
		"winner": room.G.Winner,
	})
	room.P1.Conn.Send <- msg
	room.P2.Conn.Send <- msg
}
