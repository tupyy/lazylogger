package gui

func (gui *Gui) WriteInfo(text string) {
	gui.write(INFO, text)
}

func (gui *Gui) WriteWarning(text string) {
	gui.write(WARNING, text)
}

func (gui *Gui) WriteError(text string) {
	gui.write(ERROR, text)
}

func (gui *Gui) write(level int, text string) {
	gui.debugMessages = append(gui.debugMessages, DebugMessage{Level: level, Text: text})
}
