package main

import (
	"bytes"
	"math/rand"
	"strings"
	"testing"
)

func TestCompareRLToAlphaBetaCountsAllGames(t *testing.T) {
	rlAgent := NewQLearningAgentWithSeed(8)
	rlAgent.ExplorationRate = 0

	stats := CompareRLToAlphaBetaWithOptions(rlAgent, ComparisonOptions{
		Games:          4,
		BoardSize:      5,
		AlphaBetaDepth: 1,
		RandomOpenings: 2,
		Seed:           10,
	})

	if stats.Games != 4 {
		t.Fatalf("expected 4 games, got %d", stats.Games)
	}
	if stats.RLWins+stats.AlphaBetaWins+stats.Draws != stats.Games {
		t.Fatalf("expected outcomes to add up to games, got %+v", stats)
	}
	if stats.TotalMoves == 0 {
		t.Fatal("expected comparison games to record moves")
	}
	if stats.AverageMoves() <= 0 {
		t.Fatalf("expected positive average moves, got %f", stats.AverageMoves())
	}
	if stats.OpeningMoves != 8 {
		t.Fatalf("expected 8 total opening moves, got %d", stats.OpeningMoves)
	}
	if stats.AlphaBetaDepth != 1 {
		t.Fatalf("expected alpha-beta depth 1, got %d", stats.AlphaBetaDepth)
	}
}

func TestRunCompareCommandReportsResults(t *testing.T) {
	policyPath := t.TempDir() + "/policy.json"
	agent := NewQLearningAgentWithSeed(9)
	agent.TrainSelfPlay(5, 5)
	agent.ExplorationRate = 0

	if err := agent.Save(policyPath); err != nil {
		t.Fatalf("save failed: %v", err)
	}

	var output bytes.Buffer
	err := runCompareCommand([]string{
		"-games", "2",
		"-size", "5",
		"-policy", policyPath,
		"-depth", "1",
		"-openings", "2",
		"-seed", "11",
	}, &output)
	if err != nil {
		t.Fatalf("compare command failed: %v", err)
	}

	if !strings.Contains(output.String(), "RL Policy wins") {
		t.Fatalf("expected comparison output to include results, got %q", output.String())
	}
	if !strings.Contains(output.String(), "Random opening moves: 2 | Seed: 11") {
		t.Fatalf("expected comparison output to include opening settings, got %q", output.String())
	}
}

func TestRunCompareCommandSupportsImportedMatchup(t *testing.T) {
	policyPath := t.TempDir() + "/policy.json"
	agent := NewQLearningAgentWithSeed(12)
	agent.TrainSelfPlay(5, 5)
	agent.ExplorationRate = 0

	if err := agent.Save(policyPath); err != nil {
		t.Fatalf("save failed: %v", err)
	}

	var output bytes.Buffer
	err := runCompareCommand([]string{
		"-games", "2",
		"-size", "5",
		"-policy", policyPath,
		"-depth", "1",
		"-left", "imported",
		"-right", "rl",
	}, &output)
	if err != nil {
		t.Fatalf("compare command failed: %v", err)
	}

	if !strings.Contains(output.String(), "Imported AI wins") {
		t.Fatalf("expected imported matchup output, got %q", output.String())
	}
}

func TestRunCompareCommandRejectsBadDepth(t *testing.T) {
	var output bytes.Buffer

	err := runCompareCommand([]string{"-depth", "0"}, &output)
	if err == nil {
		t.Fatal("expected invalid depth to fail")
	}
}

func TestRunCompareCommandRejectsBadOpeningCount(t *testing.T) {
	var output bytes.Buffer

	err := runCompareCommand([]string{"-openings", "-1"}, &output)
	if err == nil {
		t.Fatal("expected invalid opening count to fail")
	}
}

func TestRandomOpeningBoardIsDeterministicWithSeed(t *testing.T) {
	boardA, nextA, movesA := randomOpeningBoard(9, 4, rand.New(rand.NewSource(12)))
	boardB, nextB, movesB := randomOpeningBoard(9, 4, rand.New(rand.NewSource(12)))

	if nextA != nextB {
		t.Fatalf("expected same next player, got %c and %c", nextA, nextB)
	}
	if len(movesA) != 4 || len(movesB) != 4 {
		t.Fatalf("expected 4 opening moves, got %d and %d", len(movesA), len(movesB))
	}
	for i := range movesA {
		if movesA[i] != movesB[i] {
			t.Fatalf("expected same move at index %d, got %+v and %+v", i, movesA[i], movesB[i])
		}
	}
	if EncodeState(boardA, nextA) != EncodeState(boardB, nextB) {
		t.Fatal("expected same board state with same seed")
	}
}
