package gui

import (
	"fmt"
	"path"

	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
	"github.com/tupyy/lazylogger/conf"
)

// Menu displays a list with all the loggers available?
type Menu struct {
	*tview.Box

	// list primitive
	list *tview.List

	// Handler to be called when a new item is selected.
	handler func(logID int)
}

// NewMenu returns a new menu
func NewMenu(conf map[int]conf.LoggerConfiguration, handler func(int)) *Menu {
	l := tview.NewList()

	for k := 0; k < len(conf); k++ {
		v := conf[k]
		l.AddItem(fmt.Sprintf("%d - %s", k, v.Name), fmt.Sprintf("%s@%s: %s", v.Host.Username, v.Host.Address, path.Base(v.File)), rune('a'+k), nil)
	}

	m := Menu{
		Box:     tview.NewBox().SetBackgroundColor(tcell.ColorBlack),
		list:    l,
		handler: handler,
	}
	m.list.SetSelectedFunc(m.setSelectedItem)
	return &m
}

// Draw to screen
func (menu *Menu) Draw(screen tcell.Screen) {
	menu.Box.Draw(screen)
	x, y, width, height := menu.GetInnerRect()

	if menu.list.GetItemCount() == 0 {
		tview.Print(screen, "[red::b]No logger found.", x, y, 60, tview.AlignCenter, tcell.ColorYellow)
		tview.Print(screen, "[red::b]Please add loggers into configuration file and restart.", x, y+2, 60, tview.AlignLeft, tcell.ColorYellow)
		return
	}

	listHeight := 2 * menu.list.GetItemCount()
	if height-2 < 2*menu.list.GetItemCount() {
		menu.list.ShowSecondaryText(false)
		listHeight = menu.list.GetItemCount()
	}

	menu.list.SetRect(x, y+2, width, listHeight)

	if menu.list.HasFocus() {
		tview.Print(screen, "[:b]Please select a logger:", x, y, 50, tview.AlignLeft, tcell.ColorYellow)
	} else {
		tview.Print(screen, "Please select a logger:", x, y, 50, tview.AlignLeft, tcell.ColorGreen)
		menu.Box.SetBorder(false)
	}
	menu.list.Draw(screen)
}

// Focus delegate the focus to the list
func (menu *Menu) Focus(delegate func(p tview.Primitive)) {
	delegate(menu.list)
}

func (menu *Menu) HasFocus() bool {
	return menu.list.HasFocus()
}

func (menu *Menu) setSelectedItem(item int, main, secondary string, key rune) {
	menu.handler(item)
}
