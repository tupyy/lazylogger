package gui

import (
	"fmt"

	"github.com/rivo/tview"
)

type NavBar struct {
	*tview.TextView
}

func NewNavBar() *NavBar {
	navBar := &NavBar{TextView: tview.NewTextView().
		SetDynamicColors(true).
		SetRegions(true).
		SetWrap(false),
	}
	return navBar
}

func (navBar *NavBar) CreatePagesNavBar(names []string) {
	navBar.Clear()

	navBar.createPages(names)
}

func (navBar *NavBar) SelectPage(name string) {
	navBar.Highlight(name).ScrollToHighlight()
}

func (navBar *NavBar) createPages(names []string) {
	for i := 0; i < len(names); i++ {
		fmt.Fprintf(navBar, `%d ["%s"][darkcyan]Page %d[white][""]  `, i+1, names[i], i+1)
	}

	fmt.Fprintf(navBar, `Ctrl-H ["%s"][darkcyan]%s[white][""]  `, "help", "Help")
	fmt.Fprintf(navBar, `Ctrl-J ["%s"][darkcyan]%s[white][""]  `, "info", "Info")
}
