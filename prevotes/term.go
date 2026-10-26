package prevotes

import (
	"fmt"
	"log"
	"os"
	"time"

	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
)

// Leave enough room for the vote markers, rank, voting power and moniker.
const validatorColumnWidth = 40

func DrawScreen(network string, voteChan chan []VoteState, votePctChan, commitPctChan chan float64, summaryChan chan string) {
	if err := ui.Init(); err != nil {
		log.Fatalf("failed to initialize termui: %v", err)
	}
	votePctGauge := widgets.NewGauge()
	commitPctGauge := widgets.NewGauge()

	p := widgets.NewParagraph()
	p.Title = network

	grid := ui.NewGrid()
	termWidth, termHeight := ui.TerminalDimensions()
	var votes []VoteState
	setValidatorLayout(grid, p, votePctGauge, commitPctGauge, votes, termWidth, termHeight)
	ui.Render(grid)

	refresh := false
	tick := time.NewTicker(100 * time.Millisecond)
	defer tick.Stop()
	uiEvents := ui.PollEvents()

	for {
		select {

		case <-tick.C:
			if !refresh {
				continue
			}
			refresh = false
			ui.Render(grid)

		case e := <-uiEvents:
			switch e.ID {
			case "q", "<C-c>":
				ui.Clear()
				ui.Close()
				os.Exit(0)
			case "<Resize>":
				payload := e.Payload.(ui.Resize)
				termWidth, termHeight = payload.Width, payload.Height
				setValidatorLayout(grid, p, votePctGauge, commitPctGauge, votes, termWidth, termHeight)
				ui.Clear()
				ui.Render(grid)
			}

		case votes = <-voteChan:
			refresh = true
			setValidatorLayout(grid, p, votePctGauge, commitPctGauge, votes, termWidth, termHeight)

		case pct := <-votePctChan:
			refresh = true
			votePctGauge.Percent = int(pct * 100)

		case pct := <-commitPctChan:
			refresh = true
			commitPctGauge.Percent = int(pct * 100)

		case summary := <-summaryChan:
			refresh = true
			p.Text = summary

		}
	}
}

func setValidatorLayout(grid *ui.Grid, summary *widgets.Paragraph, voteGauge, commitGauge *widgets.Gauge, votes []VoteState, width, height int) {
	columns := validatorColumns(width, len(votes))
	groups := splitVotes(votes, columns)
	listColumns := make([]interface{}, 0, columns)
	for _, group := range groups {
		list := widgets.NewList()
		list.Border = false
		list.Rows = make([]string, len(group))
		for i, voter := range group {
			vmissing := "❌"
			if voter.Voted {
				vmissing = "✅"
			}
			if voter.VotedZeroes {
				vmissing = "🤷"
			}
			cmissing := "❌"
			if voter.Committed {
				cmissing = "✅"
			}
			list.Rows[i] = fmt.Sprintf("%-3s %-3s %s", vmissing, cmissing, voter.Description)
		}
		listColumns = append(listColumns, ui.NewCol(1/float64(columns), list))
	}

	// termui's Grid.Set appends items, so discard the previous layout first.
	grid.Items = nil
	grid.Set(
		ui.NewRow(0.1,
			ui.NewCol(1.0/3, summary),
			ui.NewCol(1.0/3, voteGauge),
			ui.NewCol(1.0/3, commitGauge),
		),
		ui.NewRow(0.9, listColumns...),
	)
	grid.SetRect(0, 0, width, height)
}

func validatorColumns(width, count int) int {
	columns := width / validatorColumnWidth
	if columns < 1 {
		columns = 1
	}
	if count > 0 && columns > count {
		columns = count
	}
	return columns
}

func splitVotes(votes []VoteState, columns int) [][]VoteState {
	groups := make([][]VoteState, columns)
	base, extra := len(votes)/columns, len(votes)%columns
	start := 0
	for i := range groups {
		length := base
		if i < extra {
			length++
		}
		groups[i] = votes[start : start+length]
		start += length
	}
	return groups
}
