package gui

import (
	"fmt"
	"path"

	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
	"github.com/tupyy/lazylogger/conf"
)

type LogView struct {
	*tview.Box
	textView            *tview.TextView
	menu                *Menu
	showMenu            bool
	selectLoggerHandler func(int, *LogView)
	conf                map[int]conf.LoggerConfiguration
	state               string
	err                 error
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

func (l *LogView) Focus(delegate func(p tview.Primitive)) {
	if l.showMenu {
		delegate(l.menu)
	} else {
		delegate(l.textView)
	}
}

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

func (l *LogView) SetTitle(host, file string) {
	l.Box.SetTitle(fmt.Sprintf(" Logging [yellow]%s [white]from host [yellow]%s ", file, host))
}

func (l *LogView) ClearTitle() {
	l.Box.SetTitle("")
}

func (l *LogView) Clear() {
	l.textView.Clear()
}

func (l *LogView) Write(data []byte) (int, error) {
	n, err := l.textView.Write(data)
	l.textView.ScrollToEnd()
	return n, err
}

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
