# Gomoku AI Project Summary

## Game Description

This project implements Gomoku on a 9x9 board. Two players, `X` and `O`, take turns placing stones on empty spaces. The first player to get five stones in a row wins. A row can be horizontal, vertical, or diagonal. If the board fills before either player gets five in a row, the game is a draw.

Gomoku is challenging for AI players because the number of possible moves is large, especially early in the game. Even on a 9x9 board, there can be many legal moves each turn, so searching every possible future game quickly becomes too expensive. The game is also tactical: a player may need to immediately block an opponent's four-in-a-row, while also trying to create their own threats. This makes short-term pattern recognition very important.

At the same time, Gomoku is easier than some other board games because the rules are simple and the win condition is easy to check. There are no hidden pieces, random events, captures, or complicated piece movement rules. Every player has full information about the board at all times.

## Alpha-Beta Evaluation Function

The alpha-beta player uses a heuristic evaluation function to score board positions when the search reaches its depth limit or a terminal state.

The evaluation function first checks for wins:

- If the AI player already has five in a row, the board receives a very large positive score: `1000000`.
- If the opponent has five in a row, the board receives a very large negative score: `-1000000`.

If neither player has won yet, the function counts runs of stones for both players. It looks in four directions: horizontal, vertical, diagonal down-right, and diagonal up-right.

The scoring weights are:

- two in a row: `10`
- three in a row: `100`
- four in a row: `1000`

The final evaluation is:

```text
AI score - opponent score
```

This means the alpha-beta player prefers boards where it has more strong lines than the opponent. Longer runs are weighted much more heavily because they are closer to becoming a winning five-in-a-row.

The alpha-beta search also uses move ordering and candidate move selection. Instead of searching every empty square, it focuses on moves near existing stones and sorts moves by their evaluation score. This makes deeper searches more practical while still considering the most relevant moves.

## Reinforcement Learning Player

The reinforcement learning player uses Q-learning. During training, it plays games against itself and updates values based on whether its moves eventually led to a win, loss, or draw.

The basic reward structure is:

- win: `+1`
- loss: `-1`
- draw: `0`

The agent uses a learning rate of `0.1`, a discount factor of `0.9`, and an exploration rate of `0.2` during training. Exploration means the agent sometimes chooses a random candidate move instead of always choosing the move it currently thinks is best. This helps it discover moves and strategies it might not otherwise try.

A simple table that memorizes exact board positions would not work very well for Gomoku, because the number of possible board states is very large. To make the RL player stronger, this project also uses tactical move features. These features help the RL player generalize across similar positions instead of only remembering exact boards.

The RL features include:

- center control
- nearby friendly stones
- nearby opponent stones
- making two, three, or four in a row
- blocking the opponent's two, three, or four in a row
- immediate winning moves
- immediate blocks against opponent wins

The RL player saves its trained policy to `rl_policy.json`. During comparison and GUI play, the policy is loaded and the exploration rate is set to `0`, so the agent plays greedily using what it learned.

The policy was trained with:

```bash
go run . train -games 3000 -size 9 -out rl_policy.json -seed 20260421
```

The training output was:

```text
Trained RL agent for 3000 games on a 9x9 board
X wins: 1370 | O wins: 956 | Draws: 674 | Invalid games: 0
Average moves: 38.09
Saved policy to rl_policy.json
```

## Experimental Comparison

To compare the RL player and alpha-beta player, I ran them against each other from randomized opening positions. I used a fixed random seed so the experiments are reproducible. The comparison alternates which player goes first.

The randomized openings are important because otherwise the same two deterministic players repeat the same game path many times. Randomized openings create a better test of how well each player handles different board situations.

The comparison commands were:

```bash
go run . compare -games 100 -size 9 -policy rl_policy.json -depth 2 -openings 4 -seed 20260422
go run . compare -games 100 -size 9 -policy rl_policy.json -depth 3 -openings 4 -seed 20260422
go run . compare -games 20 -size 9 -policy rl_policy.json -depth 4 -openings 4 -seed 20260422
```

The results were:

| Alpha-beta depth | Games | RL wins | Alpha-beta wins | Draws | Average moves |
| --- | ---: | ---: | ---: | ---: | ---: |
| 2 | 100 | 91 | 9 | 0 | 15.51 |
| 3 | 100 | 88 | 12 | 0 | 16.20 |
| 4 | 20 | 14 | 6 | 0 | 19.30 |

Depth 4 was only tested for 20 games because it takes noticeably longer to run. As alpha-beta search depth increases, the number of positions it must evaluate grows quickly.

## Conclusion

The RL player performed well against the alpha-beta player, especially at lower search depths. Against depth 2 alpha-beta, RL won 91 out of 100 games. Against depth 3, RL still won most games, but alpha-beta won more often. At depth 4, alpha-beta became even more competitive, winning 6 out of 20 games.

These results make sense because alpha-beta becomes stronger as it searches deeper. A deeper alpha-beta player can see more future threats and avoid short-term traps. However, increasing the depth also makes the program much slower because the search tree grows quickly.

The RL player's strong performance is probably due to its tactical features. It is not only memorizing exact board positions; it also recognizes useful patterns such as winning moves, blocking moves, and creating longer runs. These features help it make good decisions quickly without searching many future positions.

Overall, the experiment shows a tradeoff between the two approaches. Alpha-beta is systematic and becomes stronger with more depth, but it can become slow. The RL player is faster during play after training and performs well because it has learned and uses tactical patterns. The results do not prove that the RL player is always stronger than alpha-beta, but they do show that the trained RL player is competitive and often stronger than the alpha-beta player at the tested depths.
