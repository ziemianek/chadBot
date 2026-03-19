package twitch

import (
	"net/http"

	"github.com/charmbracelet/log"
	"github.com/gorilla/websocket"
)

const url string = "wss://eventsub.wss.twitch.tv/ws"

type Client struct {
	Conn     *websocket.Conn
	Response *http.Response
}

func (c *Client) Connect() error {
	var err error
	c.Conn, c.Response, err = websocket.DefaultDialer.Dial(url, nil)
	return err
}

func (c Client) Listen(msgChan chan string) {
	var msg []byte
	var err error
	for {
		_, msg, err = c.Conn.ReadMessage()
		if err != nil {
			log.Errorf("Twitch client could not read message: %v", err)
		}
		HandleMessage(msgChan, msg)
	}
}
