package main

import (
	"flag"
	"fmt"
	"io"
	"math/rand"
	"strings"
	"time"
)

const (
	compareAgentAlphaBeta = "alphabeta"
	compareAgentImported  = "imported"
	compareAgentRL        = "rl"
)

type ComparisonStats struct {
	Games          int
	RLWins         int
	AlphaBetaWins  int
	Draws          int
	InvalidGames   int
	TotalMoves     int
	OpeningMoves   int
	AlphaBetaDepth int
}

func (s ComparisonStats) AverageMoves() float64 {
	if s.Games == 0 {
		return 0
	}
	return float64(s.TotalMoves) / float64(s.Games)
}

type MatchupStats struct {
	Games        int
	LeftName     string
	RightName    string
	LeftWins     int
	RightWins    int
	Draws        int
	InvalidGames int
	TotalMoves   int
	OpeningMoves int
	SearchDepth  int
}

func (s MatchupStats) AverageMoves() float64 {
	if s.Games == 0 {
		return 0
	}
	return float64(s.TotalMoves) / float64(s.Games)
}

type ComparisonOptions struct {
	Games          int
	BoardSize      int
	AlphaBetaDepth int
	RandomOpenings int
	Seed           int64
}

type compareAgentSetup struct {
	agent AgentFunc
	name  string
}

func runCompareCommand(args []string, output io.Writer) error {
	flags := flag.NewFlagSet("compare", flag.ContinueOnError)
	flags.SetOutput(output)

	games := flags.Int("games", 100, "number of comparison games")
	boardSize := flags.Int("size", 9, "board size")
	policyPath := flags.String("policy", "rl_policy.json", "trained RL policy file")
	alphaBetaDepth := flags.Int("depth", 2, "search depth for alpha-beta and imported AI")
	randomOpenings := flags.Int("openings", 0, "random opening moves before agents begin")
	seed := flags.Int64("seed", time.Now().UnixNano(), "random seed for opening positions")
	leftAgentName := flags.String("left", compareAgentRL, "left-side agent: rl, alphabeta, imported")
	rightAgentName := flags.String("right", compareAgentAlphaBeta, "right-side agent: rl, alphabeta, imported")

	if err := flags.Parse(args); err != nil {
		return err
	}

	if *games <= 0 {
		return fmt.Errorf("games must be greater than 0")
	}
	if *boardSize <= 0 {
		return fmt.Errorf("size must be greater than 0")
	}
	if *alphaBetaDepth <= 0 {
		return fmt.Errorf("depth must be greater than 0")
	}
	if *randomOpenings < 0 {
		return fmt.Errorf("openings cannot be negative")
	}

	leftAgent, err := buildCompareAgent(*leftAgentName, *policyPath, *alphaBetaDepth)
	if err != nil {
		return err
	}

	rightAgent, err := buildCompareAgent(*rightAgentName, *policyPath, *alphaBetaDepth)
	if err != nil {
		return err
	}

	stats := CompareAgentsWithOptions(leftAgent.agent, rightAgent.agent, leftAgent.name, rightAgent.name, ComparisonOptions{
		Games:          *games,
		BoardSize:      *boardSize,
		AlphaBetaDepth: *alphaBetaDepth,
		RandomOpenings: *randomOpenings,
		Seed:           *seed,
	})

	fmt.Fprintf(output, "Compared %s against %s for %d games\n", stats.LeftName, stats.RightName, stats.Games)
	fmt.Fprintf(output, "Board: %dx%d | Search depth: %d\n", *boardSize, *boardSize, *alphaBetaDepth)
	fmt.Fprintf(output, "Random opening moves: %d | Seed: %d\n", *randomOpenings, *seed)
	fmt.Fprintf(output, "%s wins: %d | %s wins: %d | Draws: %d | Invalid games: %d\n", stats.LeftName, stats.LeftWins, stats.RightName, stats.RightWins, stats.Draws, stats.InvalidGames)
	fmt.Fprintf(output, "Average moves: %.2f\n", stats.AverageMoves())

	return nil
}

func buildCompareAgent(agentName string, policyPath string, depth int) (compareAgentSetup, error) {
	switch canonicalCompareAgentName(agentName) {
	case compareAgentRL:
		rlAgent, err := LoadQLearningAgent(policyPath)
		if err != nil {
			return compareAgentSetup{}, err
		}
		rlAgent.ExplorationRate = 0
		return compareAgentSetup{
			agent: rlAgent.AgentFunc(),
			name:  opponentRLPolicy,
		}, nil
	case compareAgentAlphaBeta:
		return compareAgentSetup{
			agent: AlphaBetaAgent(depth),
			name:  opponentAlphaBeta,
		}, nil
	case compareAgentImported:
		return compareAgentSetup{
			agent: ImportedAgent(depth),
			name:  opponentImportedAI,
		}, nil
	default:
		return compareAgentSetup{}, fmt.Errorf("unknown agent %q (expected rl, alphabeta, or imported)", agentName)
	}
}

func canonicalCompareAgentName(agentName string) string {
	normalized := strings.ToLower(strings.TrimSpace(agentName))
	switch normalized {
	case "alpha-beta":
		return compareAgentAlphaBeta
	case "source-ai", "source", "legacy":
		return compareAgentImported
	case "rl-policy", "policy":
		return compareAgentRL
	default:
		return normalized
	}
}

func CompareAgentsWithOptions(leftAgent AgentFunc, rightAgent AgentFunc, leftName string, rightName string, options ComparisonOptions) MatchupStats {
	stats := MatchupStats{
		Games:       options.Games,
		LeftName:    leftName,
		RightName:   rightName,
		SearchDepth: options.AlphaBetaDepth,
	}
	rng := rand.New(rand.NewSource(options.Seed))

	for game := 0; game < options.Games; game++ {
		leftPlayer := 'X'
		xAgent := leftAgent
		oAgent := rightAgent

		if game%2 == 1 {
			leftPlayer = 'O'
			xAgent = rightAgent
			oAgent = leftAgent
		}

		board, nextPlayer, openingMoves := randomOpeningBoard(options.BoardSize, options.RandomOpenings, rng)
		stats.OpeningMoves += len(openingMoves)

		result := PlayGameFromBoard(board, nextPlayer, xAgent, oAgent)
		stats.TotalMoves += len(openingMoves) + len(result.Moves)

		if result.InvalidMove {
			stats.InvalidGames++
		}

		switch result.Winner {
		case '.':
			stats.Draws++
		case leftPlayer:
			stats.LeftWins++
		default:
			stats.RightWins++
		}
	}

	return stats
}

func CompareRLToAlphaBeta(rlAgent *QLearningAgent, games int, boardSize int, alphaBetaDepth int) ComparisonStats {
	return CompareRLToAlphaBetaWithOptions(rlAgent, ComparisonOptions{
		Games:          games,
		BoardSize:      boardSize,
		AlphaBetaDepth: alphaBetaDepth,
	})
}

func CompareRLToAlphaBetaWithOptions(rlAgent *QLearningAgent, options ComparisonOptions) ComparisonStats {
	rlAgent.ExplorationRate = 0

	matchup := CompareAgentsWithOptions(
		rlAgent.AgentFunc(),
		AlphaBetaAgent(options.AlphaBetaDepth),
		opponentRLPolicy,
		opponentAlphaBeta,
		options,
	)

	return ComparisonStats{
		Games:          matchup.Games,
		RLWins:         matchup.LeftWins,
		AlphaBetaWins:  matchup.RightWins,
		Draws:          matchup.Draws,
		InvalidGames:   matchup.InvalidGames,
		TotalMoves:     matchup.TotalMoves,
		OpeningMoves:   matchup.OpeningMoves,
		AlphaBetaDepth: matchup.SearchDepth,
	}
}

func randomOpeningBoard(boardSize int, moves int, rng *rand.Rand) (Board, rune, []Move) {
	board := NewBoard(boardSize)
	player := 'X'
	openingMoves := []Move{}

	for i := 0; i < moves; i++ {
		candidates := CandidateMoves(board)
		if len(candidates) == 0 || board.Winner() != '.' {
			return board, player, openingMoves
		}

		move := candidates[rng.Intn(len(candidates))]
		if err := board.PlaceStone(move.row, move.col, player); err != nil {
			return board, player, openingMoves
		}

		openingMoves = append(openingMoves, move)
		player = otherPlayer(player)
	}

	return board, player, openingMoves
}
