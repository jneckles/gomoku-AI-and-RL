package main

func Minimax(board Board, depth int, maximizingPlayer bool, aiPlayer rune) int {
	opponent := otherPlayer(aiPlayer)

	if depth == 0 || board.HasFive(aiPlayer) || board.HasFive(opponent) || board.IsFull() {
		return Evaluate(board, aiPlayer)
	}

	moves := board.LegalMoves()

	if maximizingPlayer {
		bestScore := -1000000000

		for _, move := range moves {
			copyBoard := board.Clone()
			_ = copyBoard.PlaceStone(move.row, move.col, aiPlayer)

			score := Minimax(copyBoard, depth-1, false, aiPlayer)
			if score > bestScore {
				bestScore = score
			}
		}

		return bestScore
	}

	bestScore := 1000000000

	for _, move := range moves {
		copyBoard := board.Clone()
		_ = copyBoard.PlaceStone(move.row, move.col, opponent)

		score := Minimax(copyBoard, depth-1, true, aiPlayer)
		if score < bestScore {
			bestScore = score
		}
	}

	return bestScore
}

func BestMoveMinimax(board Board, aiPlayer rune, depth int) Move {
	moves := board.LegalMoves()
	bestMove := moves[0]
	bestScore := -1000000000

	for _, move := range moves {
		copyBoard := board.Clone()
		_ = copyBoard.PlaceStone(move.row, move.col, aiPlayer)

		score := Minimax(copyBoard, depth-1, false, aiPlayer)
		if score > bestScore {
			bestScore = score
			bestMove = move
		}
	}

	return bestMove
}
