package tui

import (
	tea "charm.land/bubbletea/v2"
	"fmt"
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
	return tea.Batch(
		m.listenForActivity(),
		m.waitForActivity(),
	)
}

func (m model) listenForActivity() tea.Cmd {
	return func() tea.Msg {
		var msg []byte
		var err error
		for {
			_, msg, err = m.client.Conn.ReadMessage()
			if err != nil {
				log.Errorf("Twitch client could not read message: %v", err)
			}
			twitch.HandleMessage(m.msgChannel, msg)
		}
	}
}

type chatMsg string

func (m model) waitForActivity() tea.Cmd {
	return func() tea.Msg {
		var msg string = <-m.msgChannel
		log.Infof("Received new message from msgChannel: %v", msg)
		return chatMsg(msg)
	}
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// for debugging purposes
	if m.dump != nil {
		spew.Fdump(m.dump, msg)
	}
	switch msg := msg.(type) {
	case chatMsg:
		m.messages = append(m.messages, string(msg))
		log.Infof("Added message to message history. Total messages: %v", len(m.messages))
		return m, m.waitForActivity()
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
	var v tea.View
	var s string = "Welcome to ChadBot\n\n"
	for _, msg := range m.messages {
		s += fmt.Sprintf("%v\n", msg)
	}
	v = tea.NewView(s)
	v.AltScreen = true
	return v
}
