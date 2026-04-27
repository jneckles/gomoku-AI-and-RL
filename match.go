package main

type AgentFunc func(board Board, player rune) Move

type GameResult struct {
	Winner        rune
	FinalBoard    Board
	Moves         []Move
	InvalidMove   bool
	InvalidPlayer rune
	InvalidReason string
}

func OnePlyAgent() AgentFunc {
	return func(board Board, player rune) Move {
		return BestMoveOnePly(board, player)
	}
}

func AlphaBetaAgent(depth int) AgentFunc {
	return func(board Board, player rune) Move {
		return BestMoveAlphaBeta(board, player, depth)
	}
}

func PlayGame(size int, xAgent AgentFunc, oAgent AgentFunc) GameResult {
	return PlayGameFromBoard(NewBoard(size), 'X', xAgent, oAgent)
}

func PlayGameFromBoard(board Board, nextPlayer rune, xAgent AgentFunc, oAgent AgentFunc) GameResult {
	result := GameResult{
		Winner:     board.Winner(),
		FinalBoard: board.Clone(),
		Moves:      []Move{},
	}

	if result.Winner != '.' || board.IsFull() {
		return result
	}

	player := nextPlayer
	for {
		agent := xAgent
		if player == 'O' {
			agent = oAgent
		}

		move := agent(board.Clone(), player)
		if err := board.PlaceStone(move.row, move.col, player); err != nil {
			result.Winner = otherPlayer(player)
			result.FinalBoard = board.Clone()
			result.InvalidMove = true
			result.InvalidPlayer = player
			result.InvalidReason = err.Error()
			return result
		}

		result.Moves = append(result.Moves, move)

		if board.HasFive(player) {
			result.Winner = player
			result.FinalBoard = board.Clone()
			return result
		}

		if board.IsFull() {
			result.Winner = '.'
			result.FinalBoard = board.Clone()
			return result
		}

		player = otherPlayer(player)
	}
}
