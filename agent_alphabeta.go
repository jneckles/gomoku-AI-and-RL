package main

import "sort"

const maxAlphaBetaMoves = 18

func AlphaBeta(board Board, depth int, alpha int, beta int, maximizingPlayer bool, aiPlayer rune) int {
	opponent := otherPlayer(aiPlayer)

	if depth == 0 || board.HasFive(aiPlayer) || board.HasFive(opponent) || board.IsFull() {
		return Evaluate(board, aiPlayer)
	}

	if maximizingPlayer {
		bestScore := -1000000000
		moves := orderedAlphaBetaMoves(board, aiPlayer, aiPlayer, true)

		for _, move := range moves {
			copyBoard := board.Clone()
			_ = copyBoard.PlaceStone(move.row, move.col, aiPlayer)

			score := AlphaBeta(copyBoard, depth-1, alpha, beta, false, aiPlayer)

			if score > bestScore {
				bestScore = score
			}
			if score > alpha {
				alpha = score
			}
			if beta <= alpha {
				break
			}
		}
		return bestScore
	}

	bestScore := 1000000000
	moves := orderedAlphaBetaMoves(board, opponent, aiPlayer, false)

	for _, move := range moves {
		copyBoard := board.Clone()
		_ = copyBoard.PlaceStone(move.row, move.col, opponent)

		score := AlphaBeta(copyBoard, depth-1, alpha, beta, true, aiPlayer)

		if score < bestScore {
			bestScore = score
		}
		if score < beta {
			beta = score
		}
		if beta <= alpha {
			break
		}
	}

	return bestScore

}

func BestMoveAlphaBeta(board Board, aiPlayer rune, depth int) Move {
	moves := orderedAlphaBetaMoves(board, aiPlayer, aiPlayer, true)
	bestMove := moves[0]
	bestScore := -1000000000
	alpha := -1000000000
	beta := 1000000000

	for _, move := range moves {
		copyBoard := board.Clone()
		_ = copyBoard.PlaceStone(move.row, move.col, aiPlayer)

		score := AlphaBeta(copyBoard, depth-1, alpha, beta, false, aiPlayer)

		if score > bestScore {
			bestScore = score
			bestMove = move
		}
		if score > alpha {
			alpha = score
		}
	}
	return bestMove
}

func orderedAlphaBetaMoves(board Board, player rune, aiPlayer rune, highScoresFirst bool) []Move {
	moves := CandidateMoves(board)
	sort.SliceStable(moves, func(i, j int) bool {
		leftBoard := board.Clone()
		rightBoard := board.Clone()

		_ = leftBoard.PlaceStone(moves[i].row, moves[i].col, player)
		_ = rightBoard.PlaceStone(moves[j].row, moves[j].col, player)

		leftScore := Evaluate(leftBoard, aiPlayer)
		rightScore := Evaluate(rightBoard, aiPlayer)

		if highScoresFirst {
			return leftScore > rightScore
		}
		return leftScore < rightScore
	})

	if len(moves) > maxAlphaBetaMoves {
		moves = moves[:maxAlphaBetaMoves]
	}

	return moves
}
