package twitch

import (
	"net/http"

	"github.com/gorilla/websocket"
)

type Client struct {
	Conn     *websocket.Conn
	Response *http.Response
}

func NewClient() *Client {
	return &Client{
		Conn:     &websocket.Conn{},
		Response: &http.Response{},
	}
}

func (c *Client) Connect(url string) error {
	var err error
	c.Conn, c.Response, err = websocket.DefaultDialer.Dial(url, nil)
	return err
}
