package gui

import (
	"github.com/gdamore/tcell"
	"github.com/tupyy/lazylogger/internal/conf"
	"github.com/tupyy/tview"
)

type views []*LogView

type LogMainView struct {
	*tview.Box

	id int

	app *tview.Application

	rootFlex *tview.Flex

	// index of the selected log view
	currentIdx int

	// holds the log views
	views views

	// keeps a copy of LoggerConfiguration to be used by the menu
	conf map[int]conf.LoggerConfiguration

	// handler called when a logger is selected in the menu.
	// the handler is passed by Gui
	selectLoggerHandler func(int, *LogView)
}

func NewLogMainView(id int, app *tview.Application, conf map[int]conf.LoggerConfiguration, selectLoggerHandler func(int, *LogView)) *LogMainView {
	l := &LogMainView{
		Box:                 tview.NewBox().SetBackgroundColor(tcell.ColorBlack),
		id:                  id,
		app:                 app,
		conf:                conf,
		selectLoggerHandler: selectLoggerHandler,
		currentIdx:          0,
		rootFlex:            tview.NewFlex(),
	}

	v := l.addView()
	l.views = append(l.views, v)

	return l
}

func (l *LogMainView) Draw(screen tcell.Screen) {
	l.Box.Draw(screen)
	l.Box.SetBorder(false)

	x, y, width, height := l.GetInnerRect()
	l.rootFlex.SetRect(x, y, width, height)
	l.rootFlex.Draw(screen)
}

// Select set focus on the current logView.
func (l *LogMainView) Select() {
	v := l.views[l.currentIdx]
	l.app.SetFocus(v)
}

func (l *LogMainView) VSplit() {
	l.rootFlex.SetDirection(tview.FlexColumn)
	v := l.addView()
	l.app.SetFocus(v)
	l.views = append(l.views, v)
}

// HSplit splits the current view horizontally
func (l *LogMainView) HSplit() {
	l.rootFlex.SetDirection(tview.FlexRow)
	v := l.addView()
	l.app.SetFocus(v)
	l.views = append(l.views, v)
}

// NextView returns the next view which can receive focus
func (l *LogMainView) NextView() {
	nextView := l.views[0]
	l.currentIdx = 0

	for idx, v := range l.views {
		if v.HasFocus() {
			if idx+1 < len(l.views) {
				nextView = l.views[idx+1]
				l.currentIdx = idx + 1
			}
		}
	}
	l.app.SetFocus(nextView)
}

func (l *LogMainView) ShowMenu() {
	v := l.getSelectedView()
	if v != nil {
		v.ClearTitle()
		v.ShowMenu()
		l.app.SetFocus(v)
	}
}

func (l *LogMainView) RemoveCurrentView() *LogView {
	v := l.getSelectedView()
	l.rootFlex.RemoveItem(v)

	idx := 0
	for i, vv := range l.views {
		if vv == v {
			idx = i
			break
		}
	}
	l.views = append(l.views[:idx], l.views[idx+1:]...)

	if len(l.views) == 0 {
		v := l.addView()
		l.views = append(l.views, v)
	}

	return v
}

func (l *LogMainView) HandleEventKey(key *tcell.EventKey) {
	if key.Key() == tcell.KeyTAB {
		l.NextView()
	} else {
		switch key.Rune() {
		case rune('v'):
			l.VSplit()
		case rune('h'):
			l.HSplit()
		case rune('m'):
			l.ShowMenu()
		case rune('x'):
			l.RemoveCurrentView()
			l.NextView()
		}
	}
}

// Return the view which has focus
func (l *LogMainView) getSelectedView() *LogView {
	for _, v := range l.views {
		if v.HasFocus() {
			return v
		}
	}
	return nil
}

func (l *LogMainView) addView() *LogView {
	view := NewLogView(l.conf, l.selectLoggerHandler)
	l.rootFlex.AddItem(view, 0, 1, true)
	return view
}
