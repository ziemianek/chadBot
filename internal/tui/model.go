package tui

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"

	"charm.land/bubbles/v2/textarea"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/log"
	"github.com/davecgh/go-spew/spew"
	"github.com/ziemianek/chadbot/internal/twitch"
)

var (
	headerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#01FAC6")).
			Bold(true).
			PaddingLeft(2)

	footerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF0000")).
			Italic(true).
			Padding(0, 1)

	subtleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241"))

	// Style for the chat box area
	chatAreaStyle = lipgloss.NewStyle().
			Padding(0, 2)

	// Style for the input box container
	inputContainerStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("#01FAC6")).
				Padding(0, 1).
				Margin(1, 0)
)

type model struct {
	client     *twitch.Client
	messages   []string
	msgChannel chan string
	dump       io.Writer
	textarea   textarea.Model

	// Track window size for layout calculations
	width  int
	height int
}

func NewModel(client *twitch.Client, dump io.Writer) model {
	ta := textarea.New()
	ta.Placeholder = "/help to list all available commands"
	ta.SetVirtualCursor(false)
	ta.ShowLineNumbers = false
	ta.Focus()
	ta.SetHeight(1)

	return model{
		client:     client,
		messages:   []string{},
		msgChannel: make(chan string),
		dump:       dump,
		textarea:   ta,
	}
}

func (m model) Init() tea.Cmd {
	if err := m.client.Connect(); err != nil {
		log.Errorf("Twitch client could not connect: %v", err)
	}
	return tea.Batch(
		m.listenForActivity(),
		m.waitForActivity(),
		textarea.Blink,
	)
}

func (m model) listenForActivity() tea.Cmd {
	return func() tea.Msg {
		var msg []byte
		var err error
		for {
			msg, err = m.client.ReadMessage()
			if err != nil {
				log.Errorf("Twitch client could not read message: %v", err)
				return nil
			}
			log.Debugf("Got new message: %v", string(msg))
			m.client.HandleMessage(m.msgChannel, msg)
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

func (m model) sendOwnerMsg(msg string) error {
	bID, err := m.client.GetBroadcasterID()
	if err != nil {
		return err
	}
	payload := struct {
		BroadcasterID string `json:"broadcaster_id"`
		SenderID      string `json:"sender_id"`
		Message       string `json:"message"`
	}{
		BroadcasterID: bID,
		SenderID:      bID,
		Message:       msg,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	resp, err := m.client.SendPost("https://api.twitch.tv/helix/chat/messages", nil, body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var respBody struct {
		Data []struct {
			MessageID string `json:"message_id"`
			IsSent    bool   `json:"is_sent"`
		} `json:"data"`
	}
	// stream data directly to struct
	// more efficient than io.ReadAll + json.Unmarshal
	if err = json.NewDecoder(resp.Body).Decode(&respBody); err != nil {
		log.Errorf("Could not unmarshal response: %v", err)
		return err
	}
	if !respBody.Data[0].IsSent {
		err := errors.New("Unexpected error. Could not send the message to the chat")
		log.Error(err)
		return err
	}
	log.Infof("successfully sent message to the chat and got response: %v", respBody)
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.dump != nil {
		spew.Fdump(m.dump, msg)
	}

	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		// Ensure textarea matches the border width
		m.textarea.SetWidth(msg.Width - 6)

	case chatMsg:
		m.messages = append(m.messages, string(msg))
		return m, m.waitForActivity()

	case tea.KeyPressMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "enter":
			ownerMsg := m.textarea.Value()
			if strings.TrimSpace(ownerMsg) == "" {
				return m, nil
			}
			if err := m.sendOwnerMsg(ownerMsg); err != nil {
				m.messages = append(m.messages, subtleStyle.Render("[SYSTEM ERROR]: Failed to send"))
			} else {
				m.messages = append(m.messages, fmt.Sprintf(lipgloss.NewStyle().Foreground(lipgloss.Color("#01FAC6")).Render("[YOU] ")+"%v", ownerMsg))
			}
			m.textarea.Reset()
			return m, nil
		}
	}

	m.textarea, cmd = m.textarea.Update(msg)
	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

func (m model) headerView() string {
	return headerStyle.Render(`
           )         (               )          
    (    ( /(   (      )\ )     (    ( /(   * )  
    )\   )\())  )\     (()/(   ( )\  )\()) )  /(  
   ((_)( (_)\((((_)(   /(_)) )((_)((_)\  ( )(_)) 
  )\___  _((_))\ _ )(_))_ ((_)_   ((_)(_(_())  
 ((/ __|| || |(_)_\(_)|   \ | _ ) / _ \|_   _|  
  | (__ | __ | / _ \  | |) || _ \| (_) | | |    
   \___||_||_|/_/ \_\ |___/ |___/ \___/  |_|`)
}

func (m model) chatView(availableHeight int) string {
	// Join all messages
	joined := strings.Join(m.messages, "\n")

	// Wrap them in the fixed-height container
	return chatAreaStyle.
		Height(availableHeight).
		MaxHeight(availableHeight).
		Render(joined)
}

func (m model) textareaView() string {
	// Wrap the textarea in a stylized border
	return inputContainerStyle.Render(m.textarea.View())
}

func (m model) footerView() string {
	return footerStyle.Render("• Press ctrl+c to quit.")
}

func (m model) View() tea.View {
	// 1. Render components that have relatively static heights
	header := m.headerView()
	divider := subtleStyle.Render(strings.Repeat("─", m.width))
	input := m.textareaView()
	footer := m.footerView()

	// 2. Calculate remaining height for chat
	// We count how many rows each rendered string takes
	occupied := lipgloss.Height(header) +
		lipgloss.Height(divider) +
		lipgloss.Height(input) +
		lipgloss.Height(footer)

	chatHeight := m.height - occupied
	if chatHeight < 0 {
		chatHeight = 0
	}

	// 3. Assemble
	s := lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		divider,
		m.chatView(chatHeight),
		input,
		footer,
	)

	// 4. Handle Cursor Positioning
	var c *tea.Cursor
	if !m.textarea.VirtualCursor() {
		c = m.textarea.Cursor()
		// Logic to find where the textarea actually is in the vertical stack
		// It is: Height(Header) + Height(Divider) + Height(ChatArea) + TopPadding of inputContainer
		c.Y = lipgloss.Height(header) + lipgloss.Height(divider) + chatHeight + 1
		// Adjust X for the border and padding
		c.X += 3
	}

	v := tea.NewView(s)
	v.AltScreen = true
	v.Cursor = c
	return v
}
