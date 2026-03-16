package tui

import (
	"fmt"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/log"
	"github.com/gorilla/websocket"
	"github.com/ziemianek/chadbot/internal/twitch"
)

type model struct {
	conn     *websocket.Conn
	messages []twitch.ChatMsg
}

func NewModel(conn *websocket.Conn) model {
	return model{
		conn:     conn,
		messages: []twitch.ChatMsg{},
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case twitch.ChatMsg:
		m.messages = append(m.messages, msg)
		log.Info(m.messages)
		// return m, nil
	case tea.KeyPressMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			//TODO: properly handle connection closing
			// m.conn.Close()
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m model) View() tea.View {
	var s string = "Welcome to ChadBot\n\n"
	for msg := range m.messages {
		s += fmt.Sprintf("User wrote message: %v\n", msg)
	}
	return tea.NewView(s)
}
