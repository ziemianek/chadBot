package twitch

import (
	"encoding/json"
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

func (m *NotificationMessage) Handle(msg []byte) error {
	var err error
	err = json.Unmarshal(msg, m)
	// log.Info(m.Payload.Event.Message.Text)
	return err
}

func HandleMessage(msg []byte) error {
	var err error
	var msgType string
	msgType, err = getMessageType(msg)
	switch msgType {
	case "session_welcome":
		var welcomeMessage WelcomeMessage
		err = welcomeMessage.Handle(msg)
	case "notification":
		var notificationMessage NotificationMessage
		err = notificationMessage.Handle(msg)
	}
	return err
}
