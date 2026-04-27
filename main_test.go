package main

import "testing"

func TestGameUIComputerMoveUsesSelectedOpponent(t *testing.T) {
	game := &GameUI{
		board:    NewBoard(9),
		opponent: opponentRLPolicy,
		rlAgent:  NewQLearningAgentWithSeed(1),
	}
	game.rlAgent.ExplorationRate = 0

	move, err := game.computerMove()
	if err != nil {
		t.Fatalf("expected RL computer move, got error: %v", err)
	}
	if move != (Move{row: 4, col: 4}) {
		t.Fatalf("expected RL policy to open in the center, got %+v", move)
	}

	game.opponent = opponentAlphaBeta
	move, err = game.computerMove()
	if err != nil {
		t.Fatalf("expected alpha-beta computer move, got error: %v", err)
	}
	if move.row < 0 || move.row >= game.board.size || move.col < 0 || move.col >= game.board.size {
		t.Fatalf("expected legal alpha-beta move, got %+v", move)
	}

	game.opponent = opponentImportedAI
	move, err = game.computerMove()
	if err != nil {
		t.Fatalf("expected imported AI move, got error: %v", err)
	}
	if move.row < 0 || move.row >= game.board.size || move.col < 0 || move.col >= game.board.size {
		t.Fatalf("expected legal imported AI move, got %+v", move)
	}
}

func TestGameUIRejectsUnknownOpponent(t *testing.T) {
	game := &GameUI{
		board:       NewBoard(9),
		opponent:    opponentAlphaBeta,
		statusLabel: nil,
	}

	game.setOpponent("Unknown")

	if game.opponent != opponentAlphaBeta {
		t.Fatalf("expected opponent to remain %s, got %s", opponentAlphaBeta, game.opponent)
	}
}
