package tui

import (
	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/log"
	"github.com/davecgh/go-spew/spew"
	"github.com/ziemianek/chadbot/internal/twitch"
	"io"
)

type model struct {
	client     twitch.Client
	messages   []string
	msgChannel chan string
	dump       io.Writer // for debugging purposes
}

func NewModel(client twitch.Client, dump io.Writer) model {
	return model{
		client:     client,
		messages:   []string{},
		msgChannel: make(chan string),
		dump:       dump,
	}
}

func (m model) Init() tea.Cmd {
	var err error
	err = m.client.Connect()
	if err != nil {
		log.Errorf("Twitch client could not connect: %v", err)
	}
	go m.client.Listen(m.msgChannel)
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// for debugging purposes
	if m.dump != nil {
		spew.Fdump(m.dump, msg)
	}
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			//TODO: properly handle connection closing
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m model) View() tea.View {
	var s string = "Welcome to ChadBot\n\n"
	return tea.NewView(s)
}
