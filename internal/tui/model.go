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

type model struct {
	client     *twitch.Client
	messages   []string
	msgChannel chan string

	// for debugging purposes
	dump io.Writer

	// message input
	textarea textarea.Model
}

func NewModel(client *twitch.Client, dump io.Writer) model {
	var ta textarea.Model = textarea.New()
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
	// for debugging purposes
	if m.dump != nil {
		spew.Fdump(m.dump, msg)
	}
	var cmd tea.Cmd
	var cmds []tea.Cmd
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
		case "enter":
			//TODO: apply some styling
			ownerMsg := m.textarea.Value()
			if err := m.sendOwnerMsg(ownerMsg); err != nil {
				m.textarea.SetValue("[ERROR]: Cant send this message!")
			}
			m.messages = append(m.messages, fmt.Sprintf("You: %v", ownerMsg))
			// send m.textarea.Value() to some channel
			// catch message and send post request to twitch api
			m.textarea.Reset()
			return m, nil
		default:
			if !m.textarea.Focused() {
				cmd = m.textarea.Focus()
				cmds = append(cmds, cmd)
			}
		}
	case tea.WindowSizeMsg:
		m.textarea.SetWidth(msg.Width)
	}

	m.textarea, cmd = m.textarea.Update(msg)
	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)

}

// TODO: APPLY SOME STYLING
func (m model) headerView() string {
	return `
           )         (              )          
   (    ( /(   (     )\ )    (   ( /(   *   )  
   )\   )\())  )\   (()/(  ( )\  )\()) )  /(  
 (((_) ((_)\((((_)(  /(_)) )((_)((_)\  ( )(_)) 
 )\___  _((_))\ _ )\(_))_ ((_)_   ((_)(_(_())  
((/ __|| || |(_)_\(_)|   \ | _ ) / _ \|_   _|  
 | (__ | __ | / _ \  | |) || _ \| (_) | | |    
  \___||_||_|/_/ \_\ |___/ |___/ \___/  |_|
`
}

// TODO: make this red or something
func (m model) footerView() string {
	return "Press crtl+c to quit."
}

func (m model) View() tea.View {
	var v tea.View
	var c *tea.Cursor
	var s string
	var offset int
	if !m.textarea.VirtualCursor() {
		c = m.textarea.Cursor()
		// Set the y offset of the cursor based on the position of the textarea
		// in the application.
		if len(m.messages) == 0 {
			offset = lipgloss.Height(m.headerView() + "\n")
		} else {
			offset = lipgloss.Height(m.headerView() + strings.Repeat("\n", len(m.messages)))
		}
		c.Y += offset
	}

	s = strings.Join([]string{
		m.headerView(),
		strings.Join(m.messages, "\n"),
		m.textarea.View(),
		m.footerView()}, "\n",
	)
	v = tea.NewView(s)
	v.AltScreen = true
	v.Cursor = c
	return v
}
