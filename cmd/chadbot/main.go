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
	for {
		var msg []byte
		var readMessageErr error
		_, msg, readMessageErr = client.Conn.ReadMessage()
		check(readMessageErr)
		log.Debug(string(msg))
		twitch.HandleMessage(msg)
	}
}
