package main

func Evaluate(board Board, aiPlayer rune) int {
	opponent := otherPlayer(aiPlayer)

	if board.HasFive(aiPlayer) {
		return 1000000
	}
	if board.HasFive(opponent) {
		return -1000000
	}

	aiScore := scorePlayer(board, aiPlayer)
	oppScore := scorePlayer(board, opponent)

	return aiScore - oppScore
}

func otherPlayer(player rune) rune {
	if player == 'X' {
		return 'O'
	}
	return 'X'
}

func scorePlayer(board Board, player rune) int {
	score := 0

	score += countRuns(board, player, 2) * 10
	score += countRuns(board, player, 3) * 100
	score += countRuns(board, player, 4) * 1000

	return score
}

func countRuns(board Board, player rune, length int) int {
	count := 0

	directions := [][2]int{
		{0, 1},
		{1, 0},
		{1, 1},
		{-1, 1},
	}

	for r := 0; r < board.size; r++ {
		for c := 0; c < board.size; c++ {
			if board.grid[r][c] != player {
				continue
			}

			for _, d := range directions {
				dr, dc := d[0], d[1]
				runLength := board.checkDirection(r, c, dr, dc, player)
				if runLength == length {
					count++
				}
			}
		}
	}

	return count
}
