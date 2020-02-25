package gui

import (
	"fmt"
	"path"

	"github.com/gdamore/tcell"
	"github.com/tupyy/tview"
	"github.com/tupyy/lazylogger/internal/conf"
)

// LogView display the content of a file. It has a status bar and a menu.
// The menu is used to select from which logger the content is displayed.
type LogView struct {
	*tview.Box

	// TextView display the data
	textView *tview.TextView

	// Menu primitive
	menu *Menu

	// If true the menu will be displayed instead of the textView
	showMenu bool

	// Handler to be called when a logger is selecte from the menu.
	selectLoggerHandler func(int, *LogView)

	// LoggerConfiguration is used by the menu.
	conf map[int]conf.LoggerConfiguration

	// Holds the health of the current selected logger.
	state string

	// Error if any of the current selected logger.
	err error
}

// NewLogText creates a new TextView primitive
func NewLogView(conf map[int]conf.LoggerConfiguration, selectLoggerHandler func(int, *LogView)) *LogView {

	l := LogView{
		Box:                 tview.NewBox().SetBackgroundColor(tcell.ColorBlack),
		textView:            tview.NewTextView(),
		showMenu:            true,
		selectLoggerHandler: selectLoggerHandler,
		conf:                conf,
	}

	menu := NewMenu(conf, l.handleMenuSelectItem)
	l.menu = menu
	return &l
}

func (l *LogView) Draw(screen tcell.Screen) {
	l.Box.Draw(screen)
	l.Box.SetBorder(true)

	x, y, width, height := l.GetInnerRect()

	if l.showMenu {
		size := len(l.conf)*2 + 6
		if size > height {
			l.menu.SetRect(x+int(width/2)-20, y, 50, height)
		} else {
			l.menu.SetRect(x+int(width/2)-20, y, 50, size)
		}
		l.menu.Draw(screen)
	} else {
		l.textView.SetRect(x, y, width, height-1)
		l.textView.Draw(screen)

		line := ""
		switch l.state {
		case "healthy":
			line = fmt.Sprintf("State: %s", ToTitle(l.state))
			line = WithPadding(line, width)
			line = fmt.Sprintf("[black:green:b]%s", line)
		case "degraded":
			line = fmt.Sprintf("State: %s. Error: %s", ToTitle(l.state), l.err.Error())
			line = WithPadding(line, width)
			line = fmt.Sprintf("[black:yellow:b]%s", line)
		case "failed":
			line = fmt.Sprintf("State: %s. Error: %s", ToTitle(l.state), l.err.Error())
			line = WithPadding(line, width)
			line = fmt.Sprintf("[black:red:b]%s", line)
		}

		tview.Print(screen, line, x, y+height-1, width, tview.AlignLeft, tcell.ColorWhite)
	}

	if l.textView.HasFocus() {
		l.Box.SetBorderColor(tcell.ColorGreen)
	} else {
		l.Box.SetBorderColor(tcell.ColorWhite)
	}
}

// Focus set the focus either on the menu is showMenu is true or on the textView.
func (l *LogView) Focus(delegate func(p tview.Primitive)) {
	if l.showMenu {
		delegate(l.menu)
	} else {
		delegate(l.textView)
	}
}

// Return true if either menu of textView has the focus.
func (logView *LogView) HasFocus() bool {
	return logView.textView.HasFocus() || logView.menu.HasFocus()
}

// ShowMenu shows the logger menu
func (logView *LogView) ShowMenu() {
	logView.showMenu = true
}

// HideMenu hides the menu and shows the defaul TextView
func (logView *LogView) HideMenu() {
	logView.showMenu = false
}

// Set the title of the box.
func (l *LogView) SetTitle(host, file string) {
	l.Box.SetTitle(fmt.Sprintf(" Logging [yellow]%s [white]from host [yellow]%s ", file, host))
}

func (l *LogView) ClearTitle() {
	l.Box.SetTitle("")
}

// Clear clears the textView.
func (l *LogView) Clear() {
	l.textView.Clear()
}

// Write writes new data to the textView.
func (l *LogView) Write(data []byte) (int, error) {
	n, err := l.textView.Write(data)
	l.textView.ScrollToEnd()
	return n, err
}

// SetState shows the state of the logger.
func (l *LogView) SetState(state string, err error) {
	l.state = state
	l.err = err
}

func (l *LogView) handleMenuSelectItem(logID int) {
	l.HideMenu()
	logger := l.conf[logID]
	l.SetTitle(logger.Host.Address, path.Base(logger.File))
	l.Clear()
	l.selectLoggerHandler(logID, l)
}
