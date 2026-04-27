package main

import "fmt"

type Move struct {
	row int
	col int
}

type Board struct {
	size int
	grid [][]rune
}

func NewBoard(size int) Board {
	var b Board
	b.size = size
	b.grid = make([][]rune, size)
	for i := 0; i < size; i++ {
		b.grid[i] = make([]rune, size)
	}
	for i := 0; i < size; i++ {
		for j := 0; j < size; j++ {
			b.grid[i][j] = '.'

		}
	}
	return b
}

func (b Board) Print() {
	for i := 0; i < b.size; i++ {
		for j := 0; j < b.size; j++ {
			fmt.Printf("%c ", b.grid[i][j])
		}
		fmt.Println()
	}
}

func (b *Board) PlaceStone(r, c int, player rune) error {
	if r < 0 || r >= b.size || c < 0 || c >= b.size {
		return fmt.Errorf("position out of bounds")
	}
	if player != 'X' && player != 'O' {
		return fmt.Errorf("invalid player")
	}
	if b.grid[r][c] != '.' {
		return fmt.Errorf("position already occupied")
	}
	b.grid[r][c] = player
	return nil
}

func (b Board) Clone() Board {
	newBoard := NewBoard(b.size)
	for i := 0; i < b.size; i++ {
		for j := 0; j < b.size; j++ {
			newBoard.grid[i][j] = b.grid[i][j]
		}
	}
	return newBoard
}

func (b Board) LegalMoves() []Move {
	moves := []Move{}
	for r := 0; r < b.size; r++ {
		for c := 0; c < b.size; c++ {
			if b.grid[r][c] == '.' {
				moves = append(moves, Move{row: r, col: c})
			}
		}
	}
	return moves
}

func (b Board) IsFull() bool {
	for r := 0; r < b.size; r++ {
		for c := 0; c < b.size; c++ {
			if b.grid[r][c] == '.' {
				return false
			}
		}
	}
	return true
}

func (b Board) Winner() rune {
	if b.HasFive('X') {
		return 'X'
	}
	if b.HasFive('O') {
		return 'O'
	}
	return '.'
}

func (b Board) HasFive(player rune) bool {
	for r := 0; r < b.size; r++ {
		for c := 0; c < b.size; c++ {
			if b.grid[r][c] != player {
				continue
			}
			// Right
			if b.checkDirection(r, c, 0, 1, player) >= 5 {
				return true
			}
			// Down
			if b.checkDirection(r, c, 1, 0, player) >= 5 {
				return true
			}
			// Down-right
			if b.checkDirection(r, c, 1, 1, player) >= 5 {
				return true
			}
			// Up-right
			if b.checkDirection(r, c, -1, 1, player) >= 5 {
				return true
			}
		}
	}
	return false
}

func (b Board) checkDirection(r, c, dr, dc int, player rune) int {
	count := 0
	for {
		// stop if out of bounds
		if r < 0 || r >= b.size || c < 0 || c >= b.size {
			break
		}
		// stop if the chain breaks
		if b.grid[r][c] != player {
			break
		}
		count++
		r += dr
		c += dc
	}
	return count
}
