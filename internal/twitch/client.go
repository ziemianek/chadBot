package twitch

import (
	"net/http"

	"github.com/gorilla/websocket"
)

type Client struct {
	conn     *websocket.Conn
	response *http.Response
}

func NewClient() *Client {
	return &Client{
		conn:     &websocket.Conn{},
		response: &http.Response{},
	}
}

func (c *Client) Connect() error {
	var err error
	var url string = "wss://eventsub.wss.twitch.tv/ws"
	c.conn, c.response, err = websocket.DefaultDialer.Dial(url, nil)
	return err
}

func (c Client) ReadMessage() ([]byte, error) {
	var msg []byte
	var err error
	_, msg, err = c.conn.ReadMessage()
	return msg, err
}

func (c Client) HandleMessage(ch chan string, msg []byte) error {
	var err error
	var msgType string
	msgType, err = getMessageType(msg)
	switch msgType {
	case "session_welcome":
		var welcomeMessage welcomeMessage
		err = welcomeMessage.handle(msg)
	case "notification":
		var notificationMessage notificationMessage
		err = notificationMessage.handle(ch, msg)
	}
	return err
}
