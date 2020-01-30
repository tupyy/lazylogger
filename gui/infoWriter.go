package gui

func (gui *Gui) WriteInfo(text string) {
	gui.writeInformation(INFO, text)
}

func (gui *Gui) WriteWarning(text string) {
	gui.writeInformation(WARNING, text)
}

func (gui *Gui) WriteError(text string) {
	gui.writeInformation(ERROR, text)
}

func (gui *Gui) writeInformation(level int, text string) {
	gui.infos = append(gui.infos, Info{Level: level, Text: text})
}
