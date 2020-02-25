package gui

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell"
	"github.com/tupyy/tview"
)

var helpTextView = tview.NewTextView()

const (
	subtitle   = `lazylogger v1.1 - Visualize logs from different hosts`
	navigation = `Right arrow: Next Page    Left arrow: Previous Page   P: Show Help     Ctrl-C: Exit`
	pages      = `Ctrl+A: Add page     Ctrl+X: Delete Page`
	window     = `v: Vertical Split     h: Hortizontal Split   m: Show Menu   x: Remove selected view`
)

func newHelpView() (content tview.Primitive) {
	// What's the size of the logo?
	lines := strings.Split(logo(), "\n")
	logoWidth := 0
	logoHeight := len(lines)
	for _, line := range lines {
		if len(line) > logoWidth {
			logoWidth = len(line)
		}
	}
	logoBox := tview.NewTextView().
		SetTextColor(tcell.ColorGreen)
	fmt.Fprint(logoBox, logo())

	// Create a frame for the subtitle and navigation infos.
	frame := tview.NewFrame(tview.NewBox()).
		SetBorders(0, 0, 0, 0, 0, 0).
		AddText(subtitle, true, tview.AlignCenter, tcell.ColorWhite).
		AddText("", true, tview.AlignCenter, tcell.ColorWhite).
		AddText(navigation, true, tview.AlignCenter, tcell.ColorDarkMagenta).
		AddText(pages, true, tview.AlignCenter, tcell.ColorDarkMagenta).
		AddText(window, true, tview.AlignCenter, tcell.ColorDarkMagenta)

	// Create a Flex layout that centers the logo and subtitle.
	flex := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(tview.NewBox(), 0, 7, false).
		AddItem(tview.NewFlex().
			AddItem(tview.NewBox(), 0, 1, false).
			AddItem(logoBox, logoWidth, 1, true).
			AddItem(tview.NewBox(), 0, 1, false), logoHeight, 1, true).
		AddItem(frame, 0, 10, false)

	return flex
}

func logo() string {
	return `
8                        8                                  
8     eeeee eeeee e    e 8     eeeee eeeee eeeee eeee eeeee 
8e    8   8 "   8 8    8 8e    8  88 8   8 8   8 8    8   8 
88    8eee8 eeee8 8eeee8 88    8   8 8e    8e    8eee 8eee8e
88    88  8 88      88   88    8   8 88 "8 88 "8 88   88   8
88eee 88  8 88ee8   88   88eee 8eee8 88ee8 88ee8 88ee 88   8



`
}
