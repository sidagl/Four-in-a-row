package server

// Player represents a connected player in the game
type Player struct {
	ID       string
	Username string
	Conn     *Connection
	Room     *Room
}
