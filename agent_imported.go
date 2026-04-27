package main

const (
	importedPosInf = 1_000_000_000
	importedNegInf = -1_000_000_000
)

func ImportedAgent(depth int) AgentFunc {
	return func(board Board, player rune) Move {
		return BestMoveImported(board, player, depth)
	}
}

func BestMoveImported(board Board, aiPlayer rune, depth int) Move {
	workingBoard := board.Clone()
	moves := CandidateMoves(workingBoard)
	if len(moves) == 0 {
		return Move{row: -1, col: -1}
	}

	bestMove := moves[0]
	bestScore := importedNegInf

	for _, move := range moves {
		if err := workingBoard.PlaceStone(move.row, move.col, aiPlayer); err != nil {
			continue
		}

		score := importedPosInf
		if !workingBoard.HasFive(aiPlayer) {
			score = importedAlphaBeta(workingBoard, depth-1, importedNegInf, importedPosInf, false, aiPlayer, move)
		}

		importedRemoveStone(&workingBoard, move)

		if score > bestScore {
			bestScore = score
			bestMove = move
		}
	}

	return bestMove
}

func importedAlphaBeta(board Board, depth int, alpha int, beta int, maximizing bool, aiPlayer rune, lastMove Move) int {
	opponent := otherPlayer(aiPlayer)
	lastPlayer := opponent
	if !maximizing {
		lastPlayer = aiPlayer
	}

	if lastMove.row != -1 && lastMove.col != -1 && board.HasFive(lastPlayer) {
		if lastPlayer == aiPlayer {
			return importedPosInf
		}
		return importedNegInf
	}

	if depth <= 0 || board.IsFull() {
		return EvaluateImported(board, aiPlayer)
	}

	moves := CandidateMoves(board)
	if len(moves) == 0 {
		return EvaluateImported(board, aiPlayer)
	}

	if maximizing {
		value := importedNegInf
		for _, move := range moves {
			if err := board.PlaceStone(move.row, move.col, aiPlayer); err != nil {
				continue
			}

			score := importedAlphaBeta(board, depth-1, alpha, beta, false, aiPlayer, move)
			importedRemoveStone(&board, move)

			if score > value {
				value = score
			}
			if value > alpha {
				alpha = value
			}
			if alpha >= beta {
				break
			}
		}
		return value
	}

	value := importedPosInf
	for _, move := range moves {
		if err := board.PlaceStone(move.row, move.col, opponent); err != nil {
			continue
		}

		score := importedAlphaBeta(board, depth-1, alpha, beta, true, aiPlayer, move)
		importedRemoveStone(&board, move)

		if score < value {
			value = score
		}
		if value < beta {
			beta = value
		}
		if beta <= alpha {
			break
		}
	}

	return value
}

func importedRemoveStone(board *Board, move Move) {
	if move.row >= 0 && move.row < board.size && move.col >= 0 && move.col < board.size {
		board.grid[move.row][move.col] = '.'
	}
}

func EvaluateImported(board Board, aiPlayer rune) int {
	opponent := otherPlayer(aiPlayer)
	score := 0
	center := board.size / 2

	for r := 0; r < board.size; r++ {
		for c := 0; c < board.size; c++ {
			switch board.grid[r][c] {
			case aiPlayer:
				score += 5 - importedManhattanDistance(r, c, center, center)
			case opponent:
				score -= 5 - importedManhattanDistance(r, c, center, center)
			}
		}
	}

	for r := 0; r < board.size; r++ {
		score += importedEvaluateLine(board.grid[r], aiPlayer)
		score -= importedEvaluateLine(board.grid[r], opponent)
	}

	for c := 0; c < board.size; c++ {
		line := make([]rune, board.size)
		for r := 0; r < board.size; r++ {
			line[r] = board.grid[r][c]
		}
		score += importedEvaluateLine(line, aiPlayer)
		score -= importedEvaluateLine(line, opponent)
	}

	for startRow := 0; startRow < board.size; startRow++ {
		line := []rune{}
		r, c := startRow, 0
		for r >= 0 && r < board.size && c >= 0 && c < board.size {
			line = append(line, board.grid[r][c])
			r++
			c++
		}
		if len(line) >= 5 {
			score += importedEvaluateLine(line, aiPlayer)
			score -= importedEvaluateLine(line, opponent)
		}
	}

	for startCol := 1; startCol < board.size; startCol++ {
		line := []rune{}
		r, c := 0, startCol
		for r >= 0 && r < board.size && c >= 0 && c < board.size {
			line = append(line, board.grid[r][c])
			r++
			c++
		}
		if len(line) >= 5 {
			score += importedEvaluateLine(line, aiPlayer)
			score -= importedEvaluateLine(line, opponent)
		}
	}

	for startRow := 0; startRow < board.size; startRow++ {
		line := []rune{}
		r, c := startRow, 0
		for r >= 0 && r < board.size && c >= 0 && c < board.size {
			line = append(line, board.grid[r][c])
			r--
			c++
		}
		if len(line) >= 5 {
			score += importedEvaluateLine(line, aiPlayer)
			score -= importedEvaluateLine(line, opponent)
		}
	}

	for startCol := 1; startCol < board.size; startCol++ {
		line := []rune{}
		r, c := board.size-1, startCol
		for r >= 0 && r < board.size && c >= 0 && c < board.size {
			line = append(line, board.grid[r][c])
			r--
			c++
		}
		if len(line) >= 5 {
			score += importedEvaluateLine(line, aiPlayer)
			score -= importedEvaluateLine(line, opponent)
		}
	}

	return score
}

func importedEvaluateLine(line []rune, player rune) int {
	score := 0
	for i := 0; i < len(line); i++ {
		if line[i] != player {
			continue
		}

		j := i
		for j < len(line) && line[j] == player {
			j++
		}

		runLength := j - i
		leftOpen := i-1 >= 0 && line[i-1] == '.'
		rightOpen := j < len(line) && line[j] == '.'
		openEnds := 0
		if leftOpen {
			openEnds++
		}
		if rightOpen {
			openEnds++
		}

		score += importedScoreRun(runLength, openEnds)
		i = j - 1
	}

	return score
}

func importedScoreRun(runLength int, openEnds int) int {
	if openEnds == 0 && runLength < 5 {
		return 0
	}

	switch runLength {
	case 5:
		return 1_000_000
	case 4:
		if openEnds == 2 {
			return 100_000
		}
		if openEnds == 1 {
			return 10_000
		}
	case 3:
		if openEnds == 2 {
			return 5_000
		}
		if openEnds == 1 {
			return 500
		}
	case 2:
		if openEnds == 2 {
			return 200
		}
		if openEnds == 1 {
			return 50
		}
	case 1:
		if openEnds == 2 {
			return 10
		}
		if openEnds == 1 {
			return 1
		}
	}

	if runLength > 5 {
		return 1_000_000
	}

	return 0
}

func importedManhattanDistance(r1 int, c1 int, r2 int, c2 int) int {
	dr := r1 - r2
	if dr < 0 {
		dr = -dr
	}

	dc := c1 - c2
	if dc < 0 {
		dc = -dc
	}

	return dr + dc
}
