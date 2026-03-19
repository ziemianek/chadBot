package tui

import (
	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/log"
	"github.com/ziemianek/chadbot/internal/twitch"
	"os"
)

func StartApp() {
	// for debugging purposes
	var dump *os.File
	var err error
	dump, err = os.OpenFile("messages.log", os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o644)
	if err != nil {
		os.Exit(1)
	}
	f, _ := os.OpenFile("log.txt", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0o644)
	log.SetOutput(f)
	log.SetFormatter(log.JSONFormatter) // Use JSON format
	// ----

	var model tea.Model = NewModel(twitch.Client{}, dump)
	var p *tea.Program = tea.NewProgram(model)
	if _, err := p.Run(); err != nil {
		log.Errorf("Could not start the app: %v", err)
		os.Exit(1)
	}
}
