package tui

import (
	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/log"
	"github.com/ziemianek/chadbot/internal/twitch"
	"os"
)

func StartApp() {
	// TODO: get rid of this
	check := func(err error) {
		if err != nil {
			// log.Errorf("Got unexpected error: %v", err)
		}
	}
	var client twitch.Client = twitch.Client{}
	var clientConnErr error
	clientConnErr = client.Connect()
	check(clientConnErr)

	var model tea.Model = NewModel(client.Conn)
	var p *tea.Program = tea.NewProgram(model)

	if _, err := p.Run(); err != nil {
		log.Errorf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}

	msgChan := make(chan string)
	go func() {
		for {
			var msg []byte
			var readMessageErr error
			_, msg, readMessageErr = client.Conn.ReadMessage()
			check(readMessageErr)
			log.Debug(string(msg))
			twitch.HandleMessage(msgChan, msg)
			log.Infof("Got new chat message: %v", <-msgChan)
		}
	}()
	for {
		twitch.ReadChatMsg(msgChan)
	}
}
