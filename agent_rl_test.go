package main

import "testing"

func TestEncodeStateIncludesPlayerAndBoard(t *testing.T) {
	board := NewBoard(9)
	xState := EncodeState(board, 'X')
	oState := EncodeState(board, 'O')

	if xState == oState {
		t.Fatal("expected state encoding to include the current player")
	}

	_ = board.PlaceStone(4, 4, 'X')
	changedState := EncodeState(board, 'X')
	if xState == changedState {
		t.Fatal("expected state encoding to change when the board changes")
	}
}

func TestQLearningAgentChoosesBestKnownLegalMove(t *testing.T) {
	board := NewBoard(9)
	agent := NewQLearningAgentWithSeed(1)
	agent.ExplorationRate = 0

	bestMove := Move{row: 4, col: 4}
	agent.SetQValue(board, 'X', Move{row: 0, col: 0}, 0.5)
	agent.SetQValue(board, 'X', bestMove, 2.0)

	move := agent.ChooseMove(board, 'X')
	if move != bestMove {
		t.Fatalf("expected move %+v, got %+v", bestMove, move)
	}
}

func TestCandidateMovesStartAtCenter(t *testing.T) {
	moves := CandidateMoves(NewBoard(9))

	if len(moves) != 1 {
		t.Fatalf("expected one opening candidate, got %d", len(moves))
	}
	if moves[0] != (Move{row: 4, col: 4}) {
		t.Fatalf("expected center opening move, got %+v", moves[0])
	}
}

func TestQLearningAgentChoosesImmediateWin(t *testing.T) {
	board := NewBoard(9)
	for c := 0; c < 4; c++ {
		_ = board.PlaceStone(4, c, 'X')
	}

	agent := NewQLearningAgentWithSeed(2)
	agent.ExplorationRate = 0

	move := agent.ChooseMove(board, 'X')
	if move != (Move{row: 4, col: 4}) {
		t.Fatalf("expected winning move at row 4 col 4, got %+v", move)
	}
}

func TestQLearningAgentBlocksImmediateWin(t *testing.T) {
	board := NewBoard(9)
	for c := 0; c < 4; c++ {
		_ = board.PlaceStone(3, c, 'O')
	}
	_ = board.PlaceStone(4, 4, 'X')

	agent := NewQLearningAgentWithSeed(2)
	agent.ExplorationRate = 0

	move := agent.ChooseMove(board, 'X')
	if move != (Move{row: 3, col: 4}) {
		t.Fatalf("expected blocking move at row 3 col 4, got %+v", move)
	}
}

func TestQLearningAgentsCanPlayCompleteGame(t *testing.T) {
	xAgent := NewQLearningAgentWithSeed(1)
	oAgent := NewQLearningAgentWithSeed(2)
	xAgent.ExplorationRate = 1
	oAgent.ExplorationRate = 1

	result := PlayGame(5, xAgent.AgentFunc(), oAgent.AgentFunc())

	if result.InvalidMove {
		t.Fatalf("expected only legal moves, got invalid move: %s", result.InvalidReason)
	}
	if len(result.Moves) == 0 {
		t.Fatal("expected at least one move")
	}
	if result.Winner != 'X' && result.Winner != 'O' && result.Winner != '.' {
		t.Fatalf("unexpected winner %c", result.Winner)
	}
	if result.Winner == '.' && !result.FinalBoard.IsFull() {
		t.Fatal("expected a draw to end with a full board")
	}
}

func TestTrainSelfPlayRecordsStatsAndLearnsQValues(t *testing.T) {
	agent := NewQLearningAgentWithSeed(3)
	agent.ExplorationRate = 1

	stats := agent.TrainSelfPlay(20, 5)

	if stats.Games != 20 {
		t.Fatalf("expected 20 games, got %d", stats.Games)
	}
	if stats.XWins+stats.OWins+stats.Draws != stats.Games {
		t.Fatalf("expected outcomes to add up to games, got %+v", stats)
	}
	if stats.TotalMoves == 0 {
		t.Fatal("expected training games to record moves")
	}
	if stats.AverageMoves() <= 0 {
		t.Fatalf("expected positive average moves, got %f", stats.AverageMoves())
	}
	if len(agent.QValues) == 0 {
		t.Fatal("expected training to create Q-values")
	}
	if len(agent.FeatureWeights) == 0 {
		t.Fatal("expected training to keep feature weights")
	}
}

func TestLearnFromGameRewardsWinnerAndPenalizesLoser(t *testing.T) {
	agent := NewQLearningAgentWithSeed(4)
	agent.LearningRate = 1
	agent.DiscountFactor = 1

	xStep := trainingStep{
		state:  "X|5|.........................",
		move:   Move{row: 0, col: 0},
		player: 'X',
	}
	oStep := trainingStep{
		state:  "O|5|X........................",
		move:   Move{row: 1, col: 0},
		player: 'O',
	}

	agent.learnFromGame([]trainingStep{xStep, oStep}, 'X')

	xValue := agent.QValues[xStep.state][moveKey(xStep.move)]
	oValue := agent.QValues[oStep.state][moveKey(oStep.move)]

	if xValue != 1 {
		t.Fatalf("expected winning X move to get 1, got %f", xValue)
	}
	if oValue != -1 {
		t.Fatalf("expected losing O move to get -1, got %f", oValue)
	}
}

func TestQLearningAgentSaveAndLoad(t *testing.T) {
	board := NewBoard(9)
	agent := NewQLearningAgentWithSeed(5)
	agent.LearningRate = 0.25
	agent.DiscountFactor = 0.75
	agent.ExplorationRate = 0
	agent.SetQValue(board, 'X', Move{row: 2, col: 3}, 1.5)
	agent.FeatureWeights["center"] = 2.5

	path := t.TempDir() + "/policy.json"
	if err := agent.Save(path); err != nil {
		t.Fatalf("save failed: %v", err)
	}

	loaded, err := LoadQLearningAgent(path)
	if err != nil {
		t.Fatalf("load failed: %v", err)
	}

	if loaded.LearningRate != agent.LearningRate {
		t.Fatalf("expected learning rate %f, got %f", agent.LearningRate, loaded.LearningRate)
	}
	if loaded.DiscountFactor != agent.DiscountFactor {
		t.Fatalf("expected discount factor %f, got %f", agent.DiscountFactor, loaded.DiscountFactor)
	}
	if loaded.ExplorationRate != agent.ExplorationRate {
		t.Fatalf("expected exploration rate %f, got %f", agent.ExplorationRate, loaded.ExplorationRate)
	}
	if loaded.FeatureWeights["center"] != 2.5 {
		t.Fatalf("expected saved center feature weight 2.5, got %f", loaded.FeatureWeights["center"])
	}

	state := EncodeState(board, 'X')
	value := loaded.QValues[state][moveKey(Move{row: 2, col: 3})]
	if value != 1.5 {
		t.Fatalf("expected saved Q-value 1.5, got %f", value)
	}
}
