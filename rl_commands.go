package main

import (
	"flag"
	"fmt"
	"io"
	"time"
)

func runTrainCommand(args []string, output io.Writer) error {
	flags := flag.NewFlagSet("train", flag.ContinueOnError)
	flags.SetOutput(output)

	games := flags.Int("games", 1000, "number of self-play games")
	boardSize := flags.Int("size", 9, "board size")
	outPath := flags.String("out", "rl_policy.json", "output policy file")
	learningRate := flags.Float64("alpha", 0.1, "learning rate")
	discountFactor := flags.Float64("gamma", 0.9, "discount factor")
	explorationRate := flags.Float64("epsilon", 0.2, "exploration rate")
	seed := flags.Int64("seed", time.Now().UnixNano(), "random seed")

	if err := flags.Parse(args); err != nil {
		return err
	}

	if *games <= 0 {
		return fmt.Errorf("games must be greater than 0")
	}
	if *boardSize <= 0 {
		return fmt.Errorf("size must be greater than 0")
	}

	agent := NewQLearningAgentWithSeed(*seed)
	agent.LearningRate = *learningRate
	agent.DiscountFactor = *discountFactor
	agent.ExplorationRate = *explorationRate

	stats := agent.TrainSelfPlay(*games, *boardSize)

	agent.ExplorationRate = 0
	if err := agent.Save(*outPath); err != nil {
		return err
	}

	fmt.Fprintf(output, "Trained RL agent for %d games on a %dx%d board\n", stats.Games, *boardSize, *boardSize)
	fmt.Fprintf(output, "X wins: %d | O wins: %d | Draws: %d | Invalid games: %d\n", stats.XWins, stats.OWins, stats.Draws, stats.InvalidGames)
	fmt.Fprintf(output, "Average moves: %.2f\n", stats.AverageMoves())
	fmt.Fprintf(output, "Saved policy to %s\n", *outPath)

	return nil
}
