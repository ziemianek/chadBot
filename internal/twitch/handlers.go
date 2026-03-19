package twitch

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/charmbracelet/log"
)

func (m *WelcomeMessage) Handle(msg []byte) error {
	var err error
	var sessionId string
	err = json.Unmarshal(msg, m)
	sessionId = m.Payload.Session.Id
	log.Infof("Extracted session id: %v", sessionId)
	subscribe(sessionId)
	return err
}

func (m *NotificationMessage) Handle(ch chan string, msg []byte) error {
	var err error
	err = json.Unmarshal(msg, m)
	// TODO: should be at debug level
	log.Infof("Parsing notification message: %v", string(msg))
	// TODO: refactor this maybe?
	ch <- fmt.Sprintf("[%v] - [%v]: %v",
		parseTimestamp(m.Metadata.MsgTimestamp),
		m.Payload.Event.ChatterUserName,
		m.Payload.Event.Message.Text,
	)
	log.Infof("Sent new message: [%v] to message channel", m.Payload.Event.Message.Text)
	return err
}

func parseTimestamp(ts string) string {
	var t time.Time
	var err error
	t, err = time.Parse(time.RFC3339Nano, ts)
	if err != nil {
		log.Errorf("error parsing timestamp: %v\n", err)
		return ""
	}
	return t.Format("15:04:05")
}

func HandleMessage(ch chan string, msg []byte) error {
	var err error
	var msgType string
	msgType, err = getMessageType(msg)
	switch msgType {
	case "session_welcome":
		var welcomeMessage WelcomeMessage
		err = welcomeMessage.Handle(msg)
	case "notification":
		var notificationMessage NotificationMessage
		err = notificationMessage.Handle(ch, msg)
	}
	return err
}
