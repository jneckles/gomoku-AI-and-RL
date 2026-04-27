package main

import (
	"encoding/json"
	"math"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"
)

type QLearningAgent struct {
	QValues         map[string]map[string]float64
	FeatureWeights  map[string]float64
	LearningRate    float64
	DiscountFactor  float64
	ExplorationRate float64

	rng *rand.Rand
}

type TrainingStats struct {
	Games        int
	XWins        int
	OWins        int
	Draws        int
	InvalidGames int
	TotalMoves   int
}

type trainingStep struct {
	state    string
	move     Move
	player   rune
	features map[string]float64
}

type savedQLearningAgent struct {
	QValues         map[string]map[string]float64 `json:"q_values"`
	FeatureWeights  map[string]float64            `json:"feature_weights"`
	LearningRate    float64                       `json:"learning_rate"`
	DiscountFactor  float64                       `json:"discount_factor"`
	ExplorationRate float64                       `json:"exploration_rate"`
}

func NewQLearningAgent() *QLearningAgent {
	return NewQLearningAgentWithSeed(time.Now().UnixNano())
}

func NewQLearningAgentWithSeed(seed int64) *QLearningAgent {
	return &QLearningAgent{
		QValues:         map[string]map[string]float64{},
		FeatureWeights:  map[string]float64{},
		LearningRate:    0.1,
		DiscountFactor:  0.9,
		ExplorationRate: 0.2,
		rng:             rand.New(rand.NewSource(seed)),
	}
}

func (a *QLearningAgent) AgentFunc() AgentFunc {
	return func(board Board, player rune) Move {
		return a.ChooseMove(board, player)
	}
}

func (a *QLearningAgent) ChooseMove(board Board, player rune) Move {
	a.ensureReady()

	moves := CandidateMoves(board)
	if len(moves) == 0 {
		return Move{row: -1, col: -1}
	}

	if a.rng.Float64() < a.ExplorationRate {
		return moves[a.rng.Intn(len(moves))]
	}

	state := EncodeState(board, player)
	return a.bestKnownMove(board, player, state, moves)
}

func (a *QLearningAgent) SetQValue(board Board, player rune, move Move, value float64) {
	a.ensureReady()

	state := EncodeState(board, player)
	if _, ok := a.QValues[state]; !ok {
		a.QValues[state] = map[string]float64{}
	}
	a.QValues[state][moveKey(move)] = value
}

func (a *QLearningAgent) TrainSelfPlay(games int, boardSize int) TrainingStats {
	a.ensureReady()

	stats := TrainingStats{Games: games}

	for game := 0; game < games; game++ {
		result, history := a.playTrainingGame(boardSize)
		stats.TotalMoves += len(result.Moves)

		if result.InvalidMove {
			stats.InvalidGames++
		}

		switch result.Winner {
		case 'X':
			stats.XWins++
		case 'O':
			stats.OWins++
		default:
			stats.Draws++
		}

		a.learnFromGame(history, result.Winner)
	}

	return stats
}

func (a *QLearningAgent) Save(path string) error {
	a.ensureReady()

	data := savedQLearningAgent{
		QValues:         a.QValues,
		FeatureWeights:  a.FeatureWeights,
		LearningRate:    a.LearningRate,
		DiscountFactor:  a.DiscountFactor,
		ExplorationRate: a.ExplorationRate,
	}

	bytes, err := json.Marshal(data)
	if err != nil {
		return err
	}
	bytes = append(bytes, '\n')

	return os.WriteFile(path, bytes, 0644)
}

func LoadQLearningAgent(path string) (*QLearningAgent, error) {
	bytes, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var data savedQLearningAgent
	if err := json.Unmarshal(bytes, &data); err != nil {
		return nil, err
	}

	agent := NewQLearningAgent()
	agent.QValues = data.QValues
	agent.FeatureWeights = data.FeatureWeights
	agent.LearningRate = data.LearningRate
	agent.DiscountFactor = data.DiscountFactor
	agent.ExplorationRate = data.ExplorationRate
	agent.ensureReady()

	return agent, nil
}

func (s TrainingStats) AverageMoves() float64 {
	if s.Games == 0 {
		return 0
	}
	return float64(s.TotalMoves) / float64(s.Games)
}

func (a *QLearningAgent) playTrainingGame(boardSize int) (GameResult, []trainingStep) {
	board := NewBoard(boardSize)
	player := 'X'
	history := []trainingStep{}
	result := GameResult{
		Winner:     '.',
		FinalBoard: board.Clone(),
		Moves:      []Move{},
	}

	for {
		state := EncodeState(board, player)
		move := a.ChooseMove(board, player)
		features := MoveFeatures(board, player, move)

		if err := board.PlaceStone(move.row, move.col, player); err != nil {
			result.Winner = otherPlayer(player)
			result.FinalBoard = board.Clone()
			result.InvalidMove = true
			result.InvalidPlayer = player
			result.InvalidReason = err.Error()
			return result, history
		}

		history = append(history, trainingStep{
			state:    state,
			move:     move,
			player:   player,
			features: features,
		})
		result.Moves = append(result.Moves, move)

		if board.HasFive(player) {
			result.Winner = player
			result.FinalBoard = board.Clone()
			return result, history
		}

		if board.IsFull() {
			result.Winner = '.'
			result.FinalBoard = board.Clone()
			return result, history
		}

		player = otherPlayer(player)
	}
}

func (a *QLearningAgent) learnFromGame(history []trainingStep, winner rune) {
	for i := len(history) - 1; i >= 0; i-- {
		step := history[i]
		reward := rewardForOutcome(step.player, winner)
		movesFromEnd := len(history) - 1 - i
		target := reward * math.Pow(a.DiscountFactor, float64(movesFromEnd))
		a.updateQValue(step.state, step.move, target)
		a.updateFeatureWeights(step.features, target)
	}
}

func (a *QLearningAgent) updateQValue(state string, move Move, target float64) {
	if _, ok := a.QValues[state]; !ok {
		a.QValues[state] = map[string]float64{}
	}

	key := moveKey(move)
	oldValue := a.QValues[state][key]
	a.QValues[state][key] = oldValue + a.LearningRate*(target-oldValue)
}

func (a *QLearningAgent) updateFeatureWeights(features map[string]float64, target float64) {
	if len(features) == 0 {
		return
	}

	prediction := a.scoreLearnedFeatures(features)
	errorValue := target - prediction
	featureLearningRate := a.LearningRate * 0.05

	for name, value := range features {
		a.FeatureWeights[name] += featureLearningRate * errorValue * value
	}
}

func rewardForOutcome(player rune, winner rune) float64 {
	if winner == '.' {
		return 0
	}
	if player == winner {
		return 1
	}
	return -1
}

func (a *QLearningAgent) bestKnownMove(board Board, player rune, state string, moves []Move) Move {
	stateValues := a.QValues[state]
	bestValue := math.Inf(-1)
	bestMoves := []Move{}

	for _, move := range moves {
		value := 0.0
		if stateValues != nil {
			value = stateValues[moveKey(move)]
		}
		value += a.scoreFeatures(MoveFeatures(board, player, move))

		if value > bestValue {
			bestValue = value
			bestMoves = []Move{move}
		} else if value == bestValue {
			bestMoves = append(bestMoves, move)
		}
	}

	return bestMoves[a.rng.Intn(len(bestMoves))]
}

func (a *QLearningAgent) scoreFeatures(features map[string]float64) float64 {
	score := 0.0
	fixedWeights := fixedFeatureWeights()
	for name, value := range features {
		score += (fixedWeights[name] + a.FeatureWeights[name]) * value
	}
	return score
}

func (a *QLearningAgent) scoreLearnedFeatures(features map[string]float64) float64 {
	score := 0.0
	for name, value := range features {
		score += a.FeatureWeights[name] * value
	}
	return score
}

func (a *QLearningAgent) ensureReady() {
	if a.QValues == nil {
		a.QValues = map[string]map[string]float64{}
	}
	if a.FeatureWeights == nil {
		a.FeatureWeights = map[string]float64{}
	}
	if a.rng == nil {
		a.rng = rand.New(rand.NewSource(time.Now().UnixNano()))
	}
}

func fixedFeatureWeights() map[string]float64 {
	return map[string]float64{
		"bias":              -0.02,
		"center":            0.15,
		"adjacent_own":      0.2,
		"adjacent_opponent": 0.05,
		"make_two":          0.25,
		"make_three":        1.0,
		"make_four":         4.0,
		"winning_move":      20.0,
		"block_two":         0.2,
		"block_three":       1.2,
		"block_four":        5.0,
		"block_win":         18.0,
	}
}

func CandidateMoves(board Board) []Move {
	legalMoves := board.LegalMoves()
	if len(legalMoves) == 0 {
		return legalMoves
	}

	hasStone := false
	seen := map[string]bool{}
	moves := []Move{}

	for r := 0; r < board.size; r++ {
		for c := 0; c < board.size; c++ {
			if board.grid[r][c] == '.' {
				continue
			}

			hasStone = true
			for dr := -2; dr <= 2; dr++ {
				for dc := -2; dc <= 2; dc++ {
					nr := r + dr
					nc := c + dc
					if nr < 0 || nr >= board.size || nc < 0 || nc >= board.size {
						continue
					}
					if board.grid[nr][nc] != '.' {
						continue
					}

					move := Move{row: nr, col: nc}
					key := moveKey(move)
					if seen[key] {
						continue
					}
					seen[key] = true
					moves = append(moves, move)
				}
			}
		}
	}

	if !hasStone {
		center := board.size / 2
		return []Move{{row: center, col: center}}
	}
	if len(moves) == 0 {
		return legalMoves
	}
	return moves
}

func MoveFeatures(board Board, player rune, move Move) map[string]float64 {
	features := map[string]float64{
		"bias": 1,
	}

	if move.row < 0 || move.row >= board.size || move.col < 0 || move.col >= board.size {
		return features
	}
	if board.grid[move.row][move.col] != '.' {
		return features
	}

	center := float64(board.size-1) / 2.0
	distance := math.Hypot(float64(move.row)-center, float64(move.col)-center)
	maxDistance := math.Hypot(center, center)
	if maxDistance > 0 {
		features["center"] = 1 - distance/maxDistance
	}

	ownAdjacent, opponentAdjacent := adjacentCounts(board, player, move)
	features["adjacent_own"] = float64(ownAdjacent) / 8.0
	features["adjacent_opponent"] = float64(opponentAdjacent) / 8.0

	addLineFeatures(features, board, player, move, "make", "winning_move")
	addLineFeatures(features, board, otherPlayer(player), move, "block", "block_win")

	return features
}

func adjacentCounts(board Board, player rune, move Move) (int, int) {
	own := 0
	opponent := otherPlayer(player)
	opp := 0

	for dr := -1; dr <= 1; dr++ {
		for dc := -1; dc <= 1; dc++ {
			if dr == 0 && dc == 0 {
				continue
			}
			r := move.row + dr
			c := move.col + dc
			if r < 0 || r >= board.size || c < 0 || c >= board.size {
				continue
			}
			if board.grid[r][c] == player {
				own++
			} else if board.grid[r][c] == opponent {
				opp++
			}
		}
	}

	return own, opp
}

func addLineFeatures(features map[string]float64, board Board, player rune, move Move, prefix string, fiveName string) {
	directions := [][2]int{
		{0, 1},
		{1, 0},
		{1, 1},
		{-1, 1},
	}

	for _, direction := range directions {
		dr := direction[0]
		dc := direction[1]
		forward, forwardOpen := countLine(board, move.row+dr, move.col+dc, dr, dc, player)
		backward, backwardOpen := countLine(board, move.row-dr, move.col-dc, -dr, -dc, player)
		length := 1 + forward + backward
		openValue := opennessValue(forwardOpen, backwardOpen)

		if length >= 5 {
			addMaxFeature(features, fiveName, 1)
		} else if length >= 4 {
			addMaxFeature(features, prefix+"_four", openValue)
		} else if length == 3 {
			addMaxFeature(features, prefix+"_three", openValue)
		} else if length == 2 {
			addMaxFeature(features, prefix+"_two", openValue)
		}
	}
}

func countLine(board Board, r int, c int, dr int, dc int, player rune) (int, bool) {
	count := 0
	for r >= 0 && r < board.size && c >= 0 && c < board.size && board.grid[r][c] == player {
		count++
		r += dr
		c += dc
	}

	open := r >= 0 && r < board.size && c >= 0 && c < board.size && board.grid[r][c] == '.'
	return count, open
}

func opennessValue(openA bool, openB bool) float64 {
	if openA && openB {
		return 1
	}
	if openA || openB {
		return 0.6
	}
	return 0.2
}

func addMaxFeature(features map[string]float64, name string, value float64) {
	if value > features[name] {
		features[name] = value
	}
}

func EncodeState(board Board, player rune) string {
	var builder strings.Builder
	builder.Grow(board.size*board.size + 8)
	builder.WriteRune(player)
	builder.WriteByte('|')
	builder.WriteString(strconv.Itoa(board.size))
	builder.WriteByte('|')

	for r := 0; r < board.size; r++ {
		for c := 0; c < board.size; c++ {
			builder.WriteRune(board.grid[r][c])
		}
	}

	return builder.String()
}

func moveKey(move Move) string {
	return strconv.Itoa(move.row) + "," + strconv.Itoa(move.col)
}
