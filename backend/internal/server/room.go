package server

import (
	"encoding/json"
	"log"
	"time"

	"github.com/google/uuid"
)

// ----------------------
// 🎮 Room & Game Structs
// ----------------------

type Room struct {
	ID  string
	P1  *Player
	P2  *Player
	G   *Game
	Hub *Hub
}

type Game struct {
	Board  [][]int // 6 rows × 7 cols
	Turn   int     // 1 = P1, 2 = P2
	Winner int     // 0 = ongoing, 1/2 = player win, 3 = draw
}

// ----------------------
// 🧩 Room Initialization
// ----------------------

func NewRoom(p1, p2 *Player) *Room {
	game := &Game{
		Board: make([][]int, 6),
		Turn:  1,
	}
	for i := range game.Board {
		game.Board[i] = make([]int, 7)
	}
	room := &Room{
		ID: uuid.NewString(),
		P1: p1,
		P2: p2,
		G:  game,
	}
	p1.Room = room
	p2.Room = room
	return room
}

// ----------------------
// 🚀 Start Game
// ----------------------

func (r *Room) StartGame() {
	log.Printf("🎮 Game started between %s and %s", r.P1.Username, r.P2.Username)

	startPayload := map[string]interface{}{
		"type":    "start",
		"players": []string{r.P1.Username, r.P2.Username},
		"time":    time.Now().Format(time.RFC3339),
	}

	// Send "start" to both players (as separate WS frames)
	if err := r.P1.Conn.Conn.WriteJSON(startPayload); err != nil {
		log.Println("⚠️ Failed to send start to P1:", err)
	}
	if err := r.P2.Conn.Conn.WriteJSON(startPayload); err != nil {
		log.Println("⚠️ Failed to send start to P2:", err)
	}

	// Immediately send initial game state
	r.broadcastState()
}

// ----------------------
// 🔄 Broadcast Game State
// ----------------------

func (r *Room) broadcastState() {
	statePayload := map[string]interface{}{
		"type":   "state",
		"gameId": r.ID,
		"board":  r.G.Board,
		"turn":   r.G.Turn,
		"winner": r.G.Winner,
	}

	if err := r.P1.Conn.Conn.WriteJSON(statePayload); err != nil {
		log.Println("⚠️ Failed to send state to P1:", err)
	}
	if err := r.P2.Conn.Conn.WriteJSON(statePayload); err != nil {
		log.Println("⚠️ Failed to send state to P2:", err)
	}
}

// ----------------------
// 🎯 Handle Player Move
// ----------------------

func (r *Room) HandleMove(player *Player, column int) {
	if r.G.Winner != 0 {
		return // Game already ended
	}

	current := r.G.Turn
	if (current == 1 && player != r.P1) || (current == 2 && player != r.P2) {
		return // Not your turn
	}

	row := r.dropDisc(column, current)
	if row == -1 {
		return // Invalid column
	}

	// Check for win or draw
	if r.checkWin(row, column, current) {
		r.G.Winner = current
		r.broadcastEnd()
		return
	}
	if r.isBoardFull() {
		r.G.Winner = 3 // Draw
		r.broadcastEnd()
		return
	}

	// Switch turns
	r.G.Turn = 3 - r.G.Turn
	r.broadcastState()
}

// ----------------------
// 🧩 Handle WS Messages
// ----------------------

func (r *Room) HandleIncomingMessage(player *Player, msg []byte) {
	var data map[string]interface{}
	if err := json.Unmarshal(msg, &data); err != nil {
		log.Println("⚠️ Invalid JSON from client:", err)
		return
	}

	if data["type"] == "move" {
		colFloat, ok := data["column"].(float64)
		if !ok {
			return
		}
		col := int(colFloat)
		r.HandleMove(player, col)
	}
}

// ----------------------
// 🧱 Core Game Mechanics
// ----------------------

func (r *Room) dropDisc(col int, player int) int {
	if col < 0 || col >= 7 {
		return -1
	}
	for row := 5; row >= 0; row-- {
		if r.G.Board[row][col] == 0 {
			r.G.Board[row][col] = player
			return row
		}
	}
	return -1
}

func (r *Room) isBoardFull() bool {
	for _, row := range r.G.Board {
		for _, cell := range row {
			if cell == 0 {
				return false
			}
		}
	}
	return true
}

func (r *Room) checkWin(row, col, player int) bool {
	directions := [][]int{
		{0, 1},  // horizontal
		{1, 0},  // vertical
		{1, 1},  // diagonal down-right
		{1, -1}, // diagonal down-left
	}
	for _, d := range directions {
		count := 1
		count += r.countDirection(row, col, d[0], d[1], player)
		count += r.countDirection(row, col, -d[0], -d[1], player)
		if count >= 4 {
			return true
		}
	}
	return false
}

func (r *Room) countDirection(row, col, dr, dc, player int) int {
	count := 0
	for {
		row += dr
		col += dc
		if row < 0 || row >= 6 || col < 0 || col >= 7 {
			break
		}
		if r.G.Board[row][col] != player {
			break
		}
		count++
	}
	return count
}

// ----------------------
// 🏁 Broadcast Game End
// ----------------------

func (r *Room) broadcastEnd() {
	winnerName := "Draw"
	if r.G.Winner == 1 {
		winnerName = r.P1.Username
	} else if r.G.Winner == 2 {
		winnerName = r.P2.Username
	}

	endPayload := map[string]interface{}{
		"type":   "end",
		"winner": winnerName,
	}

	if err := r.P1.Conn.Conn.WriteJSON(endPayload); err != nil {
		log.Println("⚠️ Failed to send end to P1:", err)
	}
	if err := r.P2.Conn.Conn.WriteJSON(endPayload); err != nil {
		log.Println("⚠️ Failed to send end to P2:", err)
	}

	log.Printf("🏁 Game ended — Winner: %s", winnerName)

	// ✅ Update leaderboard in MongoDB
	if r.G.Winner == 1 && r.Hub != nil && r.Hub.Store != nil {
		r.Hub.Store.IncrementWin(r.P1.Username)
	}
	if r.G.Winner == 2 && r.Hub != nil && r.Hub.Store != nil {
		r.Hub.Store.IncrementWin(r.P2.Username)
	}
}
