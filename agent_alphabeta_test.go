package main

import "testing"

func TestAlphaBetaChoosesImmediateWin(t *testing.T) {
	board := NewBoard(9)
	for c := 0; c < 4; c++ {
		_ = board.PlaceStone(4, c, 'O')
	}

	move := BestMoveAlphaBeta(board, 'O', 2)

	if move != (Move{row: 4, col: 4}) {
		t.Fatalf("expected alpha-beta to finish five in a row, got %+v", move)
	}
}

func TestAlphaBetaBlocksImmediateWin(t *testing.T) {
	board := NewBoard(9)
	for c := 0; c < 4; c++ {
		_ = board.PlaceStone(3, c, 'X')
	}
	_ = board.PlaceStone(4, 4, 'O')

	move := BestMoveAlphaBeta(board, 'O', 2)

	if move != (Move{row: 3, col: 4}) {
		t.Fatalf("expected alpha-beta to block five in a row, got %+v", move)
	}
}
