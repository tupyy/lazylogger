package gui

import (
	"fmt"

	"github.com/rivo/tview"
)

var infoTextView *tview.TextView

const (
	ERROR   = iota
	INFO    = iota
	WARNING = iota
)

type DebugMessage struct {
	Level int
	Text  string
}

func newInfoView() tview.Primitive {
	infoTextView = tview.NewTextView().SetDynamicColors(true)
	return infoTextView
}

func write(debugMessages []DebugMessage) {
	infoTextView.Clear()

	if len(debugMessages) == 0 {
		fmt.Fprintf(infoTextView, "No informations yet...")
	}

	for _, m := range debugMessages {
		printInfo(infoTextView, m)
	}
}

func printInfo(textView *tview.TextView, i DebugMessage) {
	switch i.Level {
	case ERROR:
		fmt.Fprintf(textView, "[red]Error[white]: %s\n", i.Text)
	case WARNING:
		fmt.Fprintf(textView, "[orange]Warning[white]: %s\n", i.Text)
	case INFO:
		fmt.Fprintf(textView, "[yellow]Info[white]: %s\n", i.Text)
	}
}
