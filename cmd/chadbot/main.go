package main

import (
	"github.com/charmbracelet/log"
	"github.com/ziemianek/chadbot/internal/twitch"
)

func main() {
	check := func(err error) {
		if err != nil {
			log.Error(err)
		}
	}
	var client twitch.Client = twitch.Client{}
	var clientConnErr error
	clientConnErr = client.Connect()
	check(clientConnErr)

	msgChan := make(chan string)
	go func() {
		for {
			var msg []byte
			var readMessageErr error
			_, msg, readMessageErr = client.Conn.ReadMessage()
			check(readMessageErr)
			log.Debug(string(msg))
			twitch.HandleMessage(msgChan, msg)
		}
	}()
	for {
		log.Infof("Got new chat message: %v", <-msgChan)
	}
}
