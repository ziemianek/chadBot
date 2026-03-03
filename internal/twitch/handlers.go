package twitch

import (
	"encoding/json"
	"github.com/charmbracelet/log"
)

type MsgHandler interface {
	Handle(msg []byte) error
}

func Handle(h MsgHandler, msg []byte) error {
	var err error
	err = h.Handle(msg)
	return err
}

func (m *WelcomeMessage) Handle(msg []byte) error {
	var err error
	err = json.Unmarshal(msg, m)
	log.Infof("Session id: %v", m.Payload.Session.Id)
	subscribe(m.Payload.Session.Id)
	return err
}

func (m *NotificationMessage) Handle(msg []byte) error {
	var err error
	err = json.Unmarshal(msg, m)
	logChatMessage(m.Payload.Event)
	return err
}

func HandleMessage(msg []byte) error {
	var message BaseMessage
	var err error
	err = json.Unmarshal(msg, &message)
	var handler MsgHandler
	switch message.Metadata.MessageType {
	case "session_welcome":
		handler = &WelcomeMessage{}
	case "notification":
		handler = &NotificationMessage{}
	default:
		return err
	}
	return handler.Handle(msg)
}
