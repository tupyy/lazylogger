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

type Info struct {
	Level int
	Text  string
}

func newInfoView() tview.Primitive {
	infoTextView = tview.NewTextView().SetDynamicColors(true)
	return infoTextView
}

func writenformations(infos []Info) {
	infoTextView.Clear()

	if len(infos) == 0 {
		fmt.Fprintf(infoTextView, "No informations yet...")
	}

	for _, info := range infos {
		printInfo(infoTextView, info)
	}
}

func printInfo(textView *tview.TextView, i Info) {
	switch i.Level {
	case ERROR:
		fmt.Fprintf(textView, "[red]Error[white]: %s\n", i.Text)
	case WARNING:
		fmt.Fprintf(textView, "[orange]Warning[white]: %s\n", i.Text)
	case INFO:
		fmt.Fprintf(textView, "[yellow]Info[white]: %s\n", i.Text)
	}
}
