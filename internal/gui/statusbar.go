package gui

import (
	"errors"

	"github.com/gdamore/tcell"
	"github.com/tupyy/tview"
)

type region struct {

	// name of the region
	name string

	// text of the region
	text string

	// foregroundColor
	foregroundColor string

	// backgroundColor
	backgroundColor string
}

type statusbar struct {
	*tview.Box

	regions map[string]region
}

func NewStatusBar() *statusbar {
	return &statusbar{
		Box:     tview.NewBox().SetBackgroundColor(tcell.ColorBlack),
		regions: make(map[string]region),
	}
}

func (s *statusbar) AddRegion(name, text, foregroundColor, backgroundColor string) error {
	if _, ok := s.regions[name]; ok {
		return errors.New("region with this name already exists")
	}

	s.regions[name] = region{name, text, foregroundColor, backgroundColor}
	return nil
}
