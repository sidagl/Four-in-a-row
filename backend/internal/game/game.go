package game

import "fmt"

const (
	Rows = 6
	Cols = 7
)

type Game struct {
	ID      string  `json:"gameId"`
	Board   [][]int `json:"board"`
	Turn    int     `json:"turn"`
	Winner  int     `json:"winner"`
	Player1 string  `json:"player1"`
	Player2 string  `json:"player2"`
}

// NewGame initializes a new empty game board
func NewGame(id, p1, p2 string) *Game {
	board := make([][]int, Rows)
	for i := range board {
		board[i] = make([]int, Cols)
	}
	return &Game{
		ID:      id,
		Board:   board,
		Turn:    1,
		Player1: p1,
		Player2: p2,
	}
}

// DropDisc inserts a disc for a player in the specified column
func (g *Game) DropDisc(player, col int) bool {
	if col < 0 || col >= Cols {
		return false
	}

	// Find first available row from bottom
	row := -1
	for r := Rows - 1; r >= 0; r-- {
		if g.Board[r][col] == 0 {
			row = r
			break
		}
	}

	// Column is full
	if row == -1 {
		return false
	}

	g.Board[row][col] = player

	// Check for win condition
	if g.CheckWin(player) {
		g.Winner = player
	} else {
		g.Turn = 3 - player // alternate turns between 1 and 2
	}

	return true
}

// CheckWin checks if the given player has 4 in a row
func (g *Game) CheckWin(player int) bool {
	// Horizontal check
	for r := 0; r < Rows; r++ {
		for c := 0; c < Cols-3; c++ {
			if g.Board[r][c] == player &&
				g.Board[r][c+1] == player &&
				g.Board[r][c+2] == player &&
				g.Board[r][c+3] == player {
				fmt.Println("Horizontal win")
				return true
			}
		}
	}

	// Vertical check
	for c := 0; c < Cols; c++ {
		for r := 0; r < Rows-3; r++ {
			if g.Board[r][c] == player &&
				g.Board[r+1][c] == player &&
				g.Board[r+2][c] == player &&
				g.Board[r+3][c] == player {
				fmt.Println("Vertical win")
				return true
			}
		}
	}

	// Diagonal down-right check
	for r := 0; r < Rows-3; r++ {
		for c := 0; c < Cols-3; c++ {
			if g.Board[r][c] == player &&
				g.Board[r+1][c+1] == player &&
				g.Board[r+2][c+2] == player &&
				g.Board[r+3][c+3] == player {
				fmt.Println("Diagonal ↘ win")
				return true
			}
		}
	}

	// Diagonal up-right check
	for r := 3; r < Rows; r++ {
		for c := 0; c < Cols-3; c++ {
			if g.Board[r][c] == player &&
				g.Board[r-1][c+1] == player &&
				g.Board[r-2][c+2] == player &&
				g.Board[r-3][c+3] == player {
				fmt.Println("Diagonal ↗ win")
				return true
			}
		}
	}

	return false
}

// PrintBoard (for debugging)
func (g *Game) PrintBoard() {
	fmt.Println("Current Board:")
	for _, row := range g.Board {
		fmt.Println(row)
	}
}
