package main

import "testing"

func scriptedAgent(moves []Move) AgentFunc {
	index := 0
	return func(board Board, player rune) Move {
		if index >= len(moves) {
			legalMoves := board.LegalMoves()
			return legalMoves[0]
		}

		move := moves[index]
		index++
		return move
	}
}

func TestPlayGameStopsWhenPlayerWins(t *testing.T) {
	xAgent := scriptedAgent([]Move{
		{row: 0, col: 0},
		{row: 0, col: 1},
		{row: 0, col: 2},
		{row: 0, col: 3},
		{row: 0, col: 4},
	})
	oAgent := scriptedAgent([]Move{
		{row: 1, col: 0},
		{row: 1, col: 1},
		{row: 1, col: 2},
		{row: 1, col: 3},
	})

	result := PlayGame(9, xAgent, oAgent)

	if result.Winner != 'X' {
		t.Fatalf("expected X to win, got %c", result.Winner)
	}
	if len(result.Moves) != 9 {
		t.Fatalf("expected 9 moves, got %d", len(result.Moves))
	}
	if !result.FinalBoard.HasFive('X') {
		t.Fatal("expected final board to contain five X stones in a row")
	}
	if result.InvalidMove {
		t.Fatalf("expected a valid game, got invalid move: %s", result.InvalidReason)
	}
}

func TestPlayGameReportsInvalidMove(t *testing.T) {
	xAgent := scriptedAgent([]Move{
		{row: 0, col: 0},
	})
	oAgent := scriptedAgent([]Move{
		{row: 0, col: 0},
	})

	result := PlayGame(9, xAgent, oAgent)

	if !result.InvalidMove {
		t.Fatal("expected invalid move")
	}
	if result.InvalidPlayer != 'O' {
		t.Fatalf("expected O to make the invalid move, got %c", result.InvalidPlayer)
	}
	if result.Winner != 'X' {
		t.Fatalf("expected X to win by invalid O move, got %c", result.Winner)
	}
}
