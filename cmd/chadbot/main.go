package main

import (
	"github.com/charmbracelet/log"
	"github.com/ziemianek/chadbot/internal"
)

func main() {
	check := func(err error) {
		if err != nil {
			log.Error(err)
		}
	}
	var client internal.Client = internal.Client{}
	clientConnErr := client.Connect()
	check(clientConnErr)
	for {
		var msg []byte
		var readMessageErr error
		_, msg, readMessageErr = client.Conn.ReadMessage()
		check(readMessageErr)
		internal.HandleMessage(msg)
	}
}
