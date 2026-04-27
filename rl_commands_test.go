package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestRunTrainCommandSavesPolicy(t *testing.T) {
	outPath := t.TempDir() + "/policy.json"
	var output bytes.Buffer

	err := runTrainCommand([]string{
		"-games", "5",
		"-size", "5",
		"-out", outPath,
		"-seed", "7",
	}, &output)
	if err != nil {
		t.Fatalf("train command failed: %v", err)
	}

	loaded, err := LoadQLearningAgent(outPath)
	if err != nil {
		t.Fatalf("expected saved policy to load: %v", err)
	}
	if len(loaded.QValues) == 0 {
		t.Fatal("expected saved policy to contain Q-values")
	}
	if loaded.ExplorationRate != 0 {
		t.Fatalf("expected saved policy to use greedy play, got epsilon %f", loaded.ExplorationRate)
	}
	if !strings.Contains(output.String(), "Saved policy") {
		t.Fatalf("expected command output to mention saved policy, got %q", output.String())
	}
}

func TestRunTrainCommandRejectsBadGameCount(t *testing.T) {
	var output bytes.Buffer

	err := runTrainCommand([]string{"-games", "0"}, &output)
	if err == nil {
		t.Fatal("expected invalid game count to fail")
	}
}
