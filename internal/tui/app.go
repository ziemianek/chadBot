package tui

import (
	"context"
	"os"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/log"
	"github.com/ziemianek/chadbot/internal/twitch"
)

type App struct {
	client  *twitch.Client
	logFile *os.File
	msgFile *os.File
}

func NewApp(client *twitch.Client, debug bool) *App {
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
		client:  client,
		logFile: logFile,
		msgFile: msgFile,
	}
}

func (a App) Run() error {
	// used to cancel the login if the user exists early
	ctx := context.Background()

	if err := a.client.Login(ctx); err != nil {
		log.Errorf("Could not authorize to Twitch: %v", err)
		return err
	}

	model := NewModel(a.client, a.msgFile)
	p := tea.NewProgram(model)

	if _, err := p.Run(); err != nil {
		log.Errorf("Could not start tea.Program: %v", err)
		return err
	}
	return nil
}
