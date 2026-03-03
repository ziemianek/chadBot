package twitch

import (
	"github.com/gorilla/websocket"
	"net/http"
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
