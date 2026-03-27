package twitch

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"sync"
	"time"

	"github.com/charmbracelet/log"
	"github.com/gorilla/websocket"
	"github.com/toqueteos/webbrowser"
)

type Client struct {
	conn  *websocket.Conn
	state string
}

func NewClient() *Client {
	var state string
	var err error
	state, err = generateState()
	if err != nil {
		log.Errorf("Could not generate state: %v", err)
		return &Client{}
	}
	return &Client{
		conn:  &websocket.Conn{},
		state: state,
	}
}

func (c *Client) setState() {
	panic("Not implemented")
}

func (c *Client) generateAuthURL() string {
	var baseURL string = "https://id.twitch.tv/oauth2/authorize"
	// Use url.Values to manage parameters safely
	var params url.Values = url.Values{}
	params.Add("response_type", "code")
	params.Add("client_id", os.Getenv("CLIENT_ID"))
	params.Add("redirect_uri", "http://localhost:3000")
	//TODO: make this a map or smth
	params.Add("scope", "user:read:chat user:write:chat user:read:email") // Space-separated
	params.Add("state", c.state)
	return baseURL + "?" + params.Encode()
}

// TODO: login only if access token is invalid
// TODO: move this to constructor
func (c *Client) Login() error {
	serverDone := &sync.WaitGroup{}
	serverDone.Add(1)
	var authUrl string = c.generateAuthURL()
	startServer(serverDone)
	if err := webbrowser.Open(authUrl); err != nil {
		return err
	}
	return nil
}

func (c *Client) Connect() error {
	var err error
	var url string = "wss://eventsub.wss.twitch.tv/ws"
	//TODO: handle response?
	c.conn, _, err = websocket.DefaultDialer.Dial(url, nil)
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

func getMessageType(msg []byte) (string, error) {
	var err error
	var envelope struct {
		Metadata struct {
			MessageType string `json:"message_type"`
		} `json:"metadata"`
	}
	err = json.Unmarshal(msg, &envelope)
	return envelope.Metadata.MessageType, err
}

func generateState() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// TODO: check if state in response matches
func startServer(wg *sync.WaitGroup) {
	var srv *http.Server = &http.Server{
		Addr:         ":3000",
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		var code string = r.URL.Query().Get("code")
		if code != "" {
			log.Infof("Login success. Auth callback code: %v", code)
			fmt.Fprintln(w, "Login success! You can close this page now.")
			// Shutdown server after successful login
			log.Info("Initiating auth callback server shutdown")
			go func() {
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				if err := srv.Shutdown(ctx); err != nil {
					fmt.Println("Server Shutdown error:", err)
				}
			}()
		} else {
			fmt.Fprintln(w, "Something went wrong")
		}
	})
	// Start server
	go func() {
		defer wg.Done()
		if err := srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			log.Errorf("ListenAndServe error: %v", err)
		}
		log.Info("Auth callback server stopped")
	}()
}
