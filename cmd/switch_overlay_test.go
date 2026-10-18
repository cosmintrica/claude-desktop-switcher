//go:build windows

package cmd

import (
	"strings"
	"testing"
)

func composeCard(t *testing.T, card *terminalCard) []string {
	t.Helper()
	rows := make([][]rune, 3)
	for index := range rows {
		rows[index] = []rune(strings.Repeat(" ", card.columns))
	}
	for _, layer := range card.layers() {
		for _, text := range layer.texts {
			for offset, character := range []rune(text.text) {
				column := text.column + offset
				if column < 0 || column >= card.columns {
					t.Fatalf("%q in row %d leaves the card at column %d of %d", text.text, text.row, column, card.columns)
				}
				rows[text.row][column] = character
			}
		}
	}
	return []string{string(rows[0]), string(rows[1]), string(rows[2])}
}

func TestTerminalCardLinesShareOneWidth(t *testing.T) {
	card := newTerminalCard("claude switch private@accounts.test", "restoring private@accounts.test", 0.55)
	for range 400 {
		card.advance()
	}

	rows := composeCard(t, card)

	if card.columns != len([]rune("$ claude switch private@accounts.test"))+1 {
		t.Fatalf("columns = %d", card.columns)
	}
	if rows[0] != "$ claude switch private@accounts.test▌" {
		t.Fatalf("command row = %q", rows[0])
	}
	bar := []rune(rows[1])
	if bar[0] != '[' || bar[len(bar)-1] != ']' {
		t.Fatalf("bar row = %q", rows[1])
	}
	inner := strings.TrimRight(strings.TrimLeft(string(bar[1:len(bar)-1]), "█"), "░")
	filled := strings.Count(rows[1], "█")
	if len([]rune(inner)) > 1 || filled < 19 || filled > 25 {
		t.Fatalf("bar after the 55%% step = %q", rows[1])
	}
	if rows[2] != "   restoring private@accounts.test    " {
		t.Fatalf("step row = %q", rows[2])
	}
}

func TestTerminalCardCentersAShortCommand(t *testing.T) {
	card := newTerminalCard("claude quit", "closing Claude", 0.1)
	for range 16 {
		card.advance()
	}

	rows := composeCard(t, card)

	if want := strings.Repeat(" ", 8) + "$ claude quit▌" + strings.Repeat(" ", 8); rows[0] != want {
		t.Fatalf("command row = %q, want %q", rows[0], want)
	}
	if want := strings.Repeat(" ", 8) + "closing Claude" + strings.Repeat(" ", 8); rows[2] != want {
		t.Fatalf("step row = %q, want %q", rows[2], want)
	}
}

func TestTerminalCardNeverMovesBack(t *testing.T) {
	card := newTerminalCard("claude switch private@accounts.test", "restoring private@accounts.test", 0.6)
	card.setStep("opening Claude", 0.3)
	if card.stepAt != 0.6 || card.target < 0.6 {
		t.Fatalf("progress went back: step at %v, target %v", card.stepAt, card.target)
	}
	for range 2000 {
		card.advance()
	}
	if card.shown > 0.6+cardCreepAhead+0.0001 {
		t.Fatalf("bar ran %v past the last step at 0.6", card.shown)
	}
	card.setStep("signed in", 1)
	for range 200 {
		card.advance()
	}
	if rows := composeCard(t, card); strings.Contains(rows[1], "░") {
		t.Fatalf("finished bar = %q", rows[1])
	}
}

func TestTerminalCardShowsAFinishedOperationAtOnce(t *testing.T) {
	rows := composeCard(t, newTerminalCard("claude account add", "account added", 1))
	if strings.TrimSpace(rows[0]) != "$ claude account add▌" || strings.Contains(rows[1], "░") {
		t.Fatalf("finished card = %q", rows)
	}
}

func TestTerminalCardShortensALongStep(t *testing.T) {
	card := newTerminalCard("claude quit", strings.Repeat("x", 40), 0.1)
	rows := composeCard(t, card)
	if !strings.HasSuffix(rows[2], "…") || len([]rune(rows[2])) != card.columns {
		t.Fatalf("step row = %q", rows[2])
	}
}

func TestPartialCell(t *testing.T) {
	tests := []struct {
		fraction float64
		eighths  bool
		want     string
	}{
		{0, true, ""},
		{0.5, true, "▌"},
		{0.99, true, "▉"},
		{0.1, false, ""},
		{0.5, false, "▒"},
		{0.8, false, "▓"},
	}
	for _, test := range tests {
		if got := partialCell(test.fraction, test.eighths); got != test.want {
			t.Errorf("partialCell(%v, %v) = %q, want %q", test.fraction, test.eighths, got, test.want)
		}
	}
}
