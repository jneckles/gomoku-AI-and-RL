package main

import (
	"fmt"
	"os"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

const (
	opponentAlphaBeta  = "Alpha-Beta"
	opponentImportedAI = "Imported AI"
	opponentRLPolicy   = "RL Policy"
)

type GameUI struct {
	board       Board
	buttons     [][]*widget.Button
	statusLabel *widget.Label
	gameOver    bool
	opponent    string
	rlAgent     *QLearningAgent
}

func main() {
	if len(os.Args) > 1 && os.Args[1] == "train" {
		if err := runTrainCommand(os.Args[2:], os.Stdout); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		return
	}

	if len(os.Args) > 1 && os.Args[1] == "compare" {
		if err := runCompareCommand(os.Args[2:], os.Stdout); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		return
	}

	runGUI()
}

func runGUI() {
	g := &GameUI{
		board:    NewBoard(9),
		gameOver: false,
		opponent: opponentAlphaBeta,
	}

	a := app.New()
	a.Settings().SetTheme(theme.DarkTheme())

	w := a.NewWindow("Gomoku AI")
	w.Resize(fyne.NewSize(900, 900))

	title := widget.NewLabel("Gomoku AI")
	title.Alignment = fyne.TextAlignCenter

	g.statusLabel = widget.NewLabel("Your turn: X")
	g.statusLabel.Alignment = fyne.TextAlignCenter

	grid := g.buildBoardUI()

	opponentSelect := widget.NewSelect([]string{opponentAlphaBeta, opponentImportedAI, opponentRLPolicy}, func(choice string) {
		g.setOpponent(choice)
	})
	opponentSelect.SetSelected(g.opponent)

	resetButton := widget.NewButton("Reset Game", func() {
		g.resetGame()
	})

	controls := container.NewHBox(
		widget.NewLabel("Computer:"),
		opponentSelect,
		resetButton,
	)

	top := container.NewVBox(
		title,
		g.statusLabel,
	)

	content := container.NewBorder(
		top,
		controls,
		nil,
		nil,
		container.NewPadded(container.NewCenter(grid)),
	)

	w.SetContent(content)
	w.ShowAndRun()
}

func (g *GameUI) buildBoardUI() *fyne.Container {
	g.buttons = make([][]*widget.Button, g.board.size)

	objects := []fyne.CanvasObject{}

	for r := 0; r < g.board.size; r++ {
		g.buttons[r] = make([]*widget.Button, g.board.size)

		for c := 0; c < g.board.size; c++ {
			row := r
			col := c

			btn := widget.NewButton(" ", func() {
				g.handlePlayerMove(row, col)
			})

			g.buttons[r][c] = btn
			objects = append(objects, btn)
		}
	}

	g.refreshBoard()
	return container.NewGridWithColumns(g.board.size, objects...)
}

func (g *GameUI) handlePlayerMove(r, c int) {
	if g.gameOver {
		return
	}

	if g.board.grid[r][c] != '.' {
		return
	}

	err := g.board.PlaceStone(r, c, 'X')
	if err != nil {
		g.statusLabel.SetText(err.Error())
		return
	}

	g.refreshBoard()

	if g.board.HasFive('X') {
		g.statusLabel.SetText("You win!")
		g.gameOver = true
		g.refreshBoard()
		return
	}

	if g.board.IsFull() {
		g.statusLabel.SetText("Draw!")
		g.gameOver = true
		g.refreshBoard()
		return
	}

	g.statusLabel.SetText("Computer is thinking...")

	move, err := g.computerMove()
	if err != nil {
		g.statusLabel.SetText(err.Error())
		return
	}

	_ = g.board.PlaceStone(move.row, move.col, 'O')
	g.refreshBoard()

	if g.board.HasFive('O') {
		g.statusLabel.SetText("Computer wins!")
		g.gameOver = true
		g.refreshBoard()
		return
	}

	if g.board.IsFull() {
		g.statusLabel.SetText("Draw!")
		g.gameOver = true
		g.refreshBoard()
		return
	}

	g.statusLabel.SetText(fmt.Sprintf("Your turn: X  |  Computer played: (%d, %d)", move.row, move.col))
}

func (g *GameUI) computerMove() (Move, error) {
	if g.opponent == opponentRLPolicy {
		if err := g.loadRLAgent(); err != nil {
			return Move{}, err
		}
		return g.rlAgent.ChooseMove(g.board, 'O'), nil
	}

	if g.opponent == opponentImportedAI {
		return BestMoveImported(g.board, 'O', 2), nil
	}

	return BestMoveAlphaBeta(g.board, 'O', 2), nil
}

func (g *GameUI) loadRLAgent() error {
	if g.rlAgent != nil {
		return nil
	}

	agent, err := LoadQLearningAgent("rl_policy.json")
	if err != nil {
		return fmt.Errorf("could not load rl_policy.json: %w", err)
	}
	agent.ExplorationRate = 0
	g.rlAgent = agent
	return nil
}

func (g *GameUI) setOpponent(opponent string) {
	if opponent != opponentAlphaBeta && opponent != opponentImportedAI && opponent != opponentRLPolicy {
		return
	}

	g.opponent = opponent
	g.resetGame()
}

func (g *GameUI) refreshBoard() {
	for r := 0; r < g.board.size; r++ {
		for c := 0; c < g.board.size; c++ {
			cell := g.board.grid[r][c]
			text := " "

			if cell == 'X' {
				text = "X"
			} else if cell == 'O' {
				text = "O"
			}

			g.buttons[r][c].SetText(text)

			if cell == 'X' {
				g.buttons[r][c].Importance = widget.DangerImportance
			} else if cell == 'O' {
				g.buttons[r][c].Importance = widget.HighImportance
			} else {
				g.buttons[r][c].Importance = widget.MediumImportance
			}

			if g.gameOver {
				g.buttons[r][c].Disable()
			} else {
				g.buttons[r][c].Enable()
			}

			g.buttons[r][c].Refresh()
		}
	}
}

func (g *GameUI) resetGame() {
	g.board = NewBoard(9)
	g.gameOver = false
	g.statusLabel.SetText(fmt.Sprintf("Your turn: X  |  Computer: %s", g.opponent))
	g.refreshBoard()
}
