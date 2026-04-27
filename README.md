# Gomoku AI

This project implements 9x9 Gomoku with human play, one-ply, minimax, alpha-beta, an imported alpha-beta variant, and a reinforcement-learning player.

## Run the GUI

```bash
go run .
```

The GUI launches a human-vs-computer game where the computer uses alpha-beta pruning.
Use the computer selector in the window to switch between alpha-beta, the imported AI, and the trained RL policy.

## Train the RL player

```bash
go run . train -games 3000 -size 9 -out rl_policy.json -seed 20260421
```

The training command uses self-play Q-learning with tactical move features and saves the learned policy as JSON.

## Compare agent matchups

```bash
go run . compare -games 100 -size 9 -policy rl_policy.json -depth 2 -openings 4 -seed 20260422
```

By default, the comparison runs `rl` vs `alphabeta`. You can also choose `imported` on either side:

```bash
go run . compare -games 50 -size 9 -policy rl_policy.json -depth 2 -left imported -right alphabeta
go run . compare -games 50 -size 9 -policy rl_policy.json -depth 2 -left imported -right rl
```

The comparison alternates which player goes first and reports wins, draws, invalid games, and average game length. Add `-openings 4 -seed 20260422` to compare from reproducible randomized opening positions instead of repeating the same empty-board game.
