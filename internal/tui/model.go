package tui

import (
	"fmt"
	"io"

	"charm.land/bubbles/v2/textarea"
	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/log"
	"github.com/davecgh/go-spew/spew"
	"github.com/ziemianek/chadbot/internal/twitch"
)

type model struct {
	client     *twitch.Client
	messages   []string
	msgChannel chan string

	// for debugging purposes
	dump io.Writer

	// message input
	// viewport viewport.Model
	textarea textarea.Model
}

func NewModel(client *twitch.Client, dump io.Writer) model {
	var ta textarea.Model = textarea.New()
	ta.Placeholder = "Once upon a time..."
	ta.SetVirtualCursor(false)
	ta.SetStyles(textarea.DefaultStyles(true)) // default to dark styles.
	ta.Focus()

	return model{
		client:     client,
		messages:   []string{},
		msgChannel: make(chan string),
		dump:       dump,
		// viewport:   vp,
		textarea: ta,
	}
}

func (m model) Init() tea.Cmd {
	var err error
	err = m.client.Connect("wss://eventsub.wss.twitch.tv/ws")
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
	case tea.WindowSizeMsg:
		m.textarea.SetWidth(msg.Width)
	}
	return m, nil
}

func (m model) View() tea.View {
	var v tea.View
	//TODO: make this header nicer
	var s string = "Welcome to ChadBot\n\n"
	for _, msg := range m.messages {
		s += fmt.Sprintf("%v\n", msg)
	}
	v = tea.NewView(s)
	v.AltScreen = true
	return v
}
