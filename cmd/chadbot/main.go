package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
	"log"
	// "net/url"
)

// Define the nested structure to match your JSON
type TwitchWelcome struct {
	Payload struct {
		Session struct {
			ID string `json:"id"`
		} `json:"session"`
	} `json:"payload"`
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// Connect to the WebSocket
	conn, _, err := websocket.DefaultDialer.Dial("wss://eventsub.wss.twitch.tv/ws", nil)
	if err != nil {
		log.Fatal("Dial error:", err)
	}
	defer conn.Close()

	// Read messages from the WebSocket
	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			log.Println("Read error:", err)
			break
		}
		fmt.Printf("Received: %s\n", msg)
		var welcome TwitchWelcome
		err1 := json.Unmarshal([]byte(msg), &welcome)
		if err1 != nil {
			log.Fatal(err1)
		}

		// Extract the ID
		sessionID := welcome.Payload.Session.ID

		fmt.Printf("Extracted Session ID: %s\n", sessionID)
	}
}
