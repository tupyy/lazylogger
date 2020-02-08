package gui

import (
	"strconv"
	"time"

	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
	"github.com/tupyy/lazylogger/log"
)

const keyOne = rune('1')

type Gui struct {
	app *tview.Application

	// Root flex which contains all other primitives
	rootFlex *tview.Flex

	// Holds a pointer to logger manager
	loggerManager *log.LoggerManager

	// Navbar shows page names
	navBar *NavBar

	// currently displayed logMainView
	currentLogMainView *LogMainView

	// Pages
	pages *tview.Pages

	// Holds log main views pointers
	views []*LogMainView

	// PageCounter is incremented on addPage and holds the last created page id.
	// It is used to avoid name clashes when pages are removed and added.
	pageCounter int

	// close the start method
	done chan interface{}
}

func NewGui(app *tview.Application, lm *log.LoggerManager) *Gui {

	gui := Gui{
		app:           app,
		loggerManager: lm,
		navBar:        NewNavBar(),
		views:         []*LogMainView{},
		pages:         tview.NewPages(),
		pageCounter:   -1,
		done:          make(chan interface{}),
	}

	return &gui
}

// Start starts a go routin which draws app every 0.5s.
// In this way, we avoid to pass app pointer to every primitive which needs to be redrawn
func (gui *Gui) Start() {
	go func(done chan interface{}) {
		for {
			select {
			case <-time.After(500 * time.Millisecond):
				gui.app.Draw()
			case <-done:
				return
			}
		}
	}(gui.done)
}

func (gui *Gui) Stop() {
	gui.done <- struct{}{}
}

// Layout returns the root flex
func (gui *Gui) Layout() tview.Primitive {
	gui.pages.AddPage("help", newHelpView(), true, true)

	gui.rootFlex = tview.NewFlex().SetDirection(tview.FlexRow).AddItem(gui.pages, 0, 1, true)
	gui.rootFlex.AddItem(gui.navBar, 1, 1, true)
	return gui.rootFlex
}

// HandleEventKey handles key events. If the key is not mapped
// to an gui actions then it is passed to the current logMainView.
func (gui *Gui) HandleEventKey(key *tcell.EventKey) {
	switch key.Key() {
	case tcell.KeyLeft:
		gui.previousPage()
	case tcell.KeyRight:
		gui.nextPage()
	case tcell.KeyCtrlA:
		gui.addPage()
	case tcell.KeyCtrlX:
		gui.removeCurrentPage()
	case tcell.KeyCtrlH:
		currentPageName, _ := gui.pages.GetFrontPage()
		if currentPageName == "help" && gui.currentLogMainView != nil {
			n := strconv.Itoa(gui.currentLogMainView.id)
			gui.pages.SwitchToPage(n)
			gui.navBar.SelectPage(n)
		} else {
			gui.showHelp()
		}
	default:
		// if the key is a page number then show the page otherwise pass the key event to the currentLogMainView.
		idx := int(key.Rune() - keyOne)
		if idx < len(gui.views) && idx >= 0 {
			gui.showPage(idx)
		} else {
			if gui.currentLogMainView != nil {
				gui.currentLogMainView.HandleEventKey(key)
			}
		}
	}
}

// When a new logger is selected using the menu, the current view
// must be unregistred from the logger currently attached to it and
// registered to the new logger
func (gui *Gui) handleLogChange(logID int, view *LogView) {
	gui.loggerManager.UnregisterWriter(view)
	gui.loggerManager.RegisterWriter(logID, view)
	gui.app.SetFocus(view)
}

func (gui *Gui) addPage() {
	gui.pageCounter++
	newLogMainView := NewLogMainView(gui.pageCounter, gui.app, gui.loggerManager.GetConfigurations(), gui.handleLogChange)
	newLogMainView.Select()

	gui.views = append(gui.views, newLogMainView)

	gui.pages.AddPage(strconv.Itoa(gui.pageCounter), newLogMainView.Layout(), true, true)

	names := make([]string, len(gui.views))
	for k, v := range gui.views {
		names[k] = strconv.Itoa(v.id)
	}
	gui.navBar.CreatePagesNavBar(names)
	gui.navBar.SelectPage(strconv.Itoa(newLogMainView.id))

	gui.currentLogMainView = newLogMainView
}

func (gui *Gui) removeCurrentPage() {
	if gui.currentLogMainView == nil {
		return
	}

	gui.pages.RemovePage(strconv.Itoa(gui.currentLogMainView.id))

	//remove current view
	idx := getIndex(gui.views, gui.currentLogMainView)
	if idx >= 0 {
		gui.views = append(gui.views[:idx], gui.views[idx+1:]...)
	}

	// delete current view and select the next one
	gui.currentLogMainView = nil
	if idx == len(gui.views) {
		idx = 0
	}

	names := make([]string, len(gui.views))
	for k, v := range gui.views {
		names[k] = strconv.Itoa(v.id)
	}

	gui.navBar.CreatePagesNavBar(names)
	if len(gui.views) == 0 {
		gui.showHelp()
		return
	}
	gui.currentLogMainView = gui.views[idx]
	gui.navBar.SelectPage(strconv.Itoa(gui.currentLogMainView.id))
}

// Show the next page. If the current page is the last page than show the first page.
// When cycling through pages, the help page is not taken into account.
func (gui *Gui) nextPage() {
	currentPageName, _ := gui.pages.GetFrontPage()
	if currentPageName == "help" {
		return
	}

	nextIdx := getIndex(gui.views, gui.currentLogMainView) + 1
	if nextIdx == len(gui.views) {
		nextIdx = 0
	}

	gui.showPage(nextIdx)
}

// Show the previous page. If the current page is the first one than show the last page.
// When cycling through pages, the help page is not taken into account.
func (gui *Gui) previousPage() {
	currentPageName, _ := gui.pages.GetFrontPage()
	if currentPageName == "help" {
		return
	}

	previousIdx := getIndex(gui.views, gui.currentLogMainView) - 1
	if previousIdx < 0 {
		previousIdx = len(gui.views) - 1
	}

	gui.showPage(previousIdx)
}

// Show page with index `idx`
func (gui *Gui) showPage(idx int) {
	gui.currentLogMainView = gui.views[idx]
	gui.currentLogMainView.Select()
	gui.pages.SwitchToPage(strconv.Itoa(gui.currentLogMainView.id))
	gui.navBar.SelectPage(strconv.Itoa(gui.currentLogMainView.id))
}

func (gui *Gui) showHelp() {
	gui.navBar.SelectPage("help")
}
