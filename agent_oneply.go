package main

func BestMoveOnePly(board Board, aiPlayer rune) Move {
	moves := board.LegalMoves()
	bestMove := moves[0]
	bestScore := -1000000

	for _, move := range moves {
		copyBoard := board.Clone()
		_ = copyBoard.PlaceStone(move.row, move.col, aiPlayer)

		score := Evaluate(copyBoard, aiPlayer)
		if score > bestScore {
			bestScore = score
			bestMove = move
		}
	}
	return bestMove
}
