package gui

import (
	"fmt"

	"github.com/rivo/tview"
)

// NavBar displays the navigation bar at the bottom of the screen.
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

// CreatePagesNavBar shows the names of the pages in the navBar.
func (navBar *NavBar) CreatePagesNavBar(names []string) {
	navBar.Clear()

	navBar.createPages(names)
}

// SelectPage highlight the page.
func (navBar *NavBar) SelectPage(name string) {
	navBar.Highlight(name).ScrollToHighlight()
}

func (navBar *NavBar) createPages(names []string) {
	for i := 0; i < len(names); i++ {
		fmt.Fprintf(navBar, `%d ["%s"][darkcyan]Page %d[white][""]  `, i+1, names[i], i+1)
	}

	fmt.Fprintf(navBar, `Ctrl-H ["%s"][darkcyan]%s[white][""]  `, "help", "Help")
}
