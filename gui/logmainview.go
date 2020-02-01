package gui

import (
	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
	"github.com/tupyy/lazylogger/conf"
)

type views []*LogView

type LogMainView struct {
	id int

	app *tview.Application

	rootFlex *tview.Flex

	currentIdx int

	views views

	conf map[int]conf.LoggerConfiguration

	selectLoggerHandler func(int, *LogView)
}

func NewLogMainView(id int, app *tview.Application, conf map[int]conf.LoggerConfiguration, selectLoggerHandler func(int, *LogView)) *LogMainView {
	logMainView := &LogMainView{
		id:                  id,
		app:                 app,
		conf:                conf,
		selectLoggerHandler: selectLoggerHandler,
		currentIdx:          0,
		rootFlex:            tview.NewFlex(),
	}

	v := logMainView.addView()
	logMainView.views = append(logMainView.views, v)

	return logMainView
}

func (logMainView *LogMainView) Layout() tview.Primitive {
	return logMainView.rootFlex
}

// Activate set focus on the current logView.
func (logMainView *LogMainView) Activate() {
	v := logMainView.views[logMainView.currentIdx]
	logMainView.app.SetFocus(v)
}

func (logMainView *LogMainView) VSplit() {
	logMainView.rootFlex.SetDirection(tview.FlexColumn)
	v := logMainView.addView()
	logMainView.app.SetFocus(v)
	logMainView.views = append(logMainView.views, v)
}

// HSplit splits the current view horizontally
func (logMainView *LogMainView) HSplit() {
	logMainView.rootFlex.SetDirection(tview.FlexRow)
	v := logMainView.addView()
	logMainView.app.SetFocus(v)
	logMainView.views = append(logMainView.views, v)
}

// GetNextView returns the next view which can receive focus
func (logMainView *LogMainView) NextView() {
	nextView := logMainView.views[0]
	logMainView.currentIdx = 0

	for idx, v := range logMainView.views {
		if v.HasFocus() {
			if idx+1 < len(logMainView.views) {
				nextView = logMainView.views[idx+1]
				logMainView.currentIdx = idx + 1
			}
		}
	}
	logMainView.app.SetFocus(nextView)
}

func (logMainView *LogMainView) ShowMenu() {
	v := logMainView.getSelectedView()
	if v != nil {
		v.ClearTitle()
		v.ShowMenu()
		logMainView.app.SetFocus(v)
	}
}

func (logMainView *LogMainView) RemoveCurrentView() *LogView {
	v := logMainView.getSelectedView()
	logMainView.rootFlex.RemoveItem(v)

	idx := 0
	for i, vv := range logMainView.views {
		if vv == v {
			idx = i
			break
		}
	}
	logMainView.views = append(logMainView.views[:idx], logMainView.views[idx+1:]...)

	if len(logMainView.views) == 0 {
		v := logMainView.addView()
		logMainView.views = append(logMainView.views, v)
	}

	return v
}

func (logMainView *LogMainView) HandleEventKey(key *tcell.EventKey) {
	if key.Key() == tcell.KeyTAB {
		logMainView.NextView()
	} else {
		switch key.Rune() {
		case rune('v'):
			logMainView.VSplit()
		case rune('h'):
			logMainView.HSplit()
		case rune('m'):
			logMainView.ShowMenu()
		case rune('x'):
			logMainView.RemoveCurrentView()
			logMainView.NextView()
		}
	}
}

func (logMainView *LogMainView) getSelectedView() *LogView {

	for _, v := range logMainView.views {
		if v.HasFocus() {
			return v
		}
	}
	return nil
}

func (logMainView *LogMainView) addView() *LogView {
	view := NewLogView(logMainView.conf, logMainView.selectLoggerHandler)
	logMainView.rootFlex.AddItem(view, 0, 1, true)
	return view
}
