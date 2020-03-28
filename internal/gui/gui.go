package gui

import (
	"strconv"
	"time"

	"github.com/gdamore/tcell"
	"github.com/tupyy/lazylogger/internal/log"
	"github.com/tupyy/tview"
)

const (
	keyOne       = rune('1')
	helpPageName = "help"
)

type Gui struct {
	app *tview.Application

	// Root flex which contains all other primitives
	rootFlex *tview.Flex

	// Holds a pointer to logger manager
	loggerManager *log.LoggerManager

	// Navbar shows page names
	navBar *NavBar

	// Pages
	pages *tview.Pages

	// index of the last visible page before switching to help page
	lastIndex int

	// holds the page names
	pageNames []string

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
		pages:         tview.NewPages(),
		pageCounter:   -1,
		lastIndex:     -1,
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
	gui.pages.AddPage(helpPageName, newHelpView(), true, true)

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
		name, _ := gui.pages.GetFrontPage()
		if name != helpPageName {
			gui.removePage(name)
		}
	case tcell.KeyCtrlH:
		if gui.pages.GetPageCount() == 1 {
			return
		}
		name, _ := gui.pages.GetFrontPage()
		if name == helpPageName && gui.lastIndex > -1 {
			gui.pages.SwitchToPage(gui.pageNames[gui.lastIndex])
			gui.navBar.SelectPage(gui.pageNames[gui.lastIndex])
		} else {
			gui.lastIndex = getIndex(gui.pageNames, name)
			gui.showHelp()
		}
	default:
		// if the key is a page number then show the page otherwise pass the key event to the currentLogMainView.
		idx := int(key.Rune() - keyOne)
		if idx < len(gui.pageNames) && idx >= 0 {
			gui.showPage(idx)
		} else {
			if gui.currentLogMainView() != nil {
				gui.currentLogMainView().HandleEventKey(key)
			}
		}
	}
}

// When a new logger is selected using the menu, the current view
// must be unregistred from the logger currently attached to it and
// registered to the new logger
func (gui *Gui) handleLogChange(logID int, view *LogView) {
	gui.loggerManager.UnregisterWriter(logID)
	err := gui.loggerManager.RegisterWriter(logID, view)
	if err != nil {
		view.SetState("failed", err)
	}
	gui.app.SetFocus(view)
}

func (gui *Gui) addPage() {

	// keep incrementing page counter to avoid clashes when pages are deleted
	gui.pageCounter++
	newLogMainView := NewLogMainView(gui.pageCounter, gui.app, gui.loggerManager.GetConfigurations(), gui.handleLogChange)
	newLogMainView.Select()

	gui.pages.AddPage(strconv.Itoa(gui.pageCounter), newLogMainView, true, true)

	gui.pageNames = append(gui.pageNames, strconv.Itoa(gui.pageCounter))
	gui.navBar.CreatePagesNavBar(gui.pageNames)
	gui.navBar.SelectPage(strconv.Itoa(newLogMainView.id))
}

func (gui *Gui) removePage(name string) {
	if !gui.pages.HasPage(name) {
		return
	}

	gui.pages.RemovePage(name)
	for idx, n := range gui.pageNames {
		if n == name {
			gui.pageNames = append(gui.pageNames[:idx], gui.pageNames[idx+1:]...)
			if len(gui.pageNames) == 0 {
				gui.navBar.CreatePagesNavBar(gui.pageNames)
				return
			}
			break
		}
	}

	n, _ := gui.pages.GetFrontPage()
	gui.lastIndex = getIndex(gui.pageNames, n)
	gui.navBar.CreatePagesNavBar(gui.pageNames)
	gui.navBar.SelectPage(n)
}

// Show the next page. If the current page is the last page than show the first page.
// When cycling through pages, the help page is not taken into account.
func (gui *Gui) nextPage() {
	name, _ := gui.pages.GetFrontPage()
	if name == helpPageName {
		return
	}

	nextIdx := getIndex(gui.pageNames, name) + 1
	if nextIdx == len(gui.pageNames) {
		nextIdx = 0
	}

	gui.showPage(nextIdx)
}

// Show the previous page. If the current page is the first one than show the last page.
// When cycling through pages, the help page is not taken into account.
func (gui *Gui) previousPage() {
	name, _ := gui.pages.GetFrontPage()
	if name == helpPageName {
		return
	}

	previousIdx := getIndex(gui.pageNames, name) - 1
	if previousIdx < 0 {
		previousIdx = len(gui.pageNames) - 1
	}
	gui.showPage(previousIdx)
}

func (gui *Gui) currentLogMainView() *LogMainView {
	_, p := gui.pages.GetFrontPage()
	if l, ok := p.(*LogMainView); ok {
		return l
	}

	return nil
}

// Show page with index `idx`
func (gui *Gui) showPage(idx int) {
	name := gui.pageNames[idx]
	gui.pages.SwitchToPage(name)
	gui.currentLogMainView().Select()
	gui.navBar.SelectPage(name)
	gui.lastIndex = idx
}

func (gui *Gui) showHelp() {
	gui.pages.SwitchToPage(helpPageName)
	gui.navBar.SelectPage(helpPageName)
}
