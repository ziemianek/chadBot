package tui

import (
	"os"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/log"
	"github.com/ziemianek/chadbot/internal/twitch"
)

type App struct {
	logFile *os.File
	msgFile *os.File
}

func NewApp(debug bool) *App {
	logFile, err := os.OpenFile("build/log.txt", os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o644)
	if err != nil {
		log.Errorf("Got error while creating log file: %v", err)
	}
	log.SetOutput(logFile)
	log.SetFormatter(log.JSONFormatter)
	if debug {
		log.SetLevel(log.DebugLevel)
	}

	msgFile, err := os.OpenFile("build/messages.log", os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o644)
	if err != nil {
		log.Errorf("Got error while creating bubbletea messages history file: %v", err)
	}

	return &App{
		logFile: logFile,
		msgFile: msgFile,
	}
}

func (a App) Run() error {
	var client *twitch.Client = twitch.NewClient()
	var model tea.Model = NewModel(client, a.msgFile)
	var p *tea.Program = tea.NewProgram(model)
	if _, err := p.Run(); err != nil {
		log.Errorf("Error while running tea program: %v", err)
		return err
	}
	return nil
}
