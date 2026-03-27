package twitch

// TODO: OMG REFACTOR THIS SPAGHETTI GIANT

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/charmbracelet/log"
	"github.com/gorilla/websocket"
	"github.com/toqueteos/webbrowser"
)

const (
	TwitchAuthURL    = "https://id.twitch.tv/oauth2/authorize"
	TwitchTokenURL   = "https://id.twitch.tv/oauth2/token"
	TwitchEventsubWS = "wss://eventsub.wss.twitch.tv/ws"
	RedirectURI      = "http://localhost:3000"
)

func NewClient(repo SecretsRepository) (*Client, error) {
	id, secret, err := repo.GetCredentials()
	if err != nil {
		return nil, err
	}
	state, _ := generateState()
	return &Client{
		secretsRepo:  repo,
		clientID:     id,
		clientSecret: secret,
		state:        state,
	}, nil
}

type Client struct {
	secretsRepo  SecretsRepository
	clientID     string
	clientSecret string
	state        string
	accessToken  *TwitchToken
	conn         *websocket.Conn
}

func (c *Client) Login(ctx context.Context) error {
	if savedToken, err := c.secretsRepo.GetToken(ctx); err == nil {
		c.accessToken = savedToken
		// TODO: check if token is expired here and refresh if needed
		return nil
	}

	// if no token, do the browser flow
	resChan := make(chan string)
	srv := startServer(resChan)
	defer srv.Shutdown(ctx)

	webbrowser.Open(getAuthURL(c.clientID, c.state))

	select {
	case <-ctx.Done():
		return ctx.Err()
	case code := <-resChan:
		token, err := c.generateToken(code)
		if err != nil {
			return err
		}
		c.accessToken = token
		return c.secretsRepo.SaveToken(ctx, token)
	}
}

func (c *Client) Connect() error {
	var err error
	c.conn, _, err = websocket.DefaultDialer.Dial(TwitchEventsubWS, nil)
	return err
}

func (c *Client) ReadMessage() ([]byte, error) {
	_, msg, err := c.conn.ReadMessage()
	return msg, err
}

// HandleMessage decodes the message type and routes it to specific logic.
func (c *Client) HandleMessage(ch chan string, msg []byte) error {
	msgType, err := getMessageType(msg)
	if err != nil {
		return err
	}

	switch msgType {
	case "session_welcome":
		var m welcomeMessage
		if err := json.Unmarshal(msg, &m); err != nil {
			return err
		}
		log.Infof("WebSocket Session ID: %s", m.Payload.Session.Id)
		return c.subscribe(m.Payload.Session.Id)

	case "notification":
		var m notificationMessage
		if err := json.Unmarshal(msg, &m); err != nil {
			return err
		}

		formatted := fmt.Sprintf("[%s] %s: %s",
			parseTimestamp(m.Metadata.MsgTimestamp),
			m.Payload.Event.ChatterUserName,
			m.Payload.Event.Message.Text,
		)
		ch <- formatted
		return nil
	}
	return nil
}

func (c *Client) generateToken(code string) (*TwitchToken, error) {
	data := url.Values{}
	data.Set("client_id", c.clientID)
	data.Set("client_secret", c.clientSecret)
	data.Set("code", code)
	data.Set("grant_type", "authorization_code")
	data.Set("redirect_uri", RedirectURI)

	resp, err := http.Post(TwitchTokenURL, "application/x-www-form-urlencoded", strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var t TwitchToken
	json.NewDecoder(resp.Body).Decode(&t)
	return &t, nil
}

func (c *Client) callAPI(method, url string, headers RequestHeaders, body []byte) (*http.Response, error) {
	req, err := http.NewRequest(method, url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	if c.accessToken == nil || c.accessToken.AccessToken == "" {
		return nil, errors.New("no valid access token")
	}

	req.Header.Set("Authorization", "Bearer "+c.accessToken.AccessToken)
	req.Header.Set("Client-Id", c.clientID)
	// default to JSON if not specified, move this from subscribe to here
	if method == http.MethodPost {
		req.Header.Set("Content-Type", "application/json")
	}
	// add any additional headers
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	return http.DefaultClient.Do(req)
}

func (c *Client) SendGet(url string) (*http.Response, error) {
	return c.callAPI(http.MethodGet, url, nil, nil)
}

func (c *Client) SendPost(url string, headers RequestHeaders, body []byte) (*http.Response, error) {
	return c.callAPI(http.MethodPost, url, headers, body)
}

func (c *Client) subscribe(sessionId string) error {
	broadcasterID, err := c.getBroadcasterID()
	if err != nil {
		return fmt.Errorf("failed to get broadcaster id: %w", err)
	}

	payload := payloadSubscribeToChat{
		Type:    "channel.chat.message",
		Version: "1",
		Condition: map[string]string{
			"broadcaster_user_id": broadcasterID,
			"user_id":             broadcasterID,
		},
		Transport: map[string]string{
			"method":     "websocket",
			"session_id": sessionId,
		},
	}

	body, _ := json.Marshal(payload)
	resp, err := c.SendPost("https://api.twitch.tv/helix/eventsub/subscriptions", nil, body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted {
		out, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("subscription failed (%d): %s", resp.StatusCode, string(out))
	}

	log.Info("Successfully subscribed to chat")
	return nil
}

func (c *Client) getBroadcasterID() (string, error) {
	resp, err := c.SendGet("https://api.twitch.tv/helix/users")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("twitch api error: %s", resp.Status)
	}

	var respBody struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	// stream data directly to struct
	// more efficient than io.ReadAll + json.Unmarshal
	if err = json.NewDecoder(resp.Body).Decode(&respBody); err != nil {
		log.Errorf("Could not unmarshal response: %v", err)
		return "", err
	}

	if len(respBody.Data) == 0 {
		return "", errors.New("no user found for this token")
	}

	id := respBody.Data[0].ID
	log.Infof("Fetched Broadcaster User ID: %s", id)
	return id, nil
}

func generateState() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func getAuthURL(clientID, state string) string {
	// Use url.Values to manage parameters safely
	var params url.Values = url.Values{}
	params.Add("response_type", "code")
	params.Add("client_id", clientID)
	params.Add("redirect_uri", RedirectURI)
	//TODO: refactor scope
	params.Add("scope", "user:read:chat user:write:chat user:read:email")
	params.Add("state", state)
	return TwitchAuthURL + "?" + params.Encode()
}

func getMessageType(msg []byte) (string, error) {
	var envelope struct {
		Metadata struct {
			MessageType string `json:"message_type"`
		} `json:"metadata"`
	}
	err := json.Unmarshal(msg, &envelope)
	return envelope.Metadata.MessageType, err
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

// TODO: check if state in response matches
func startServer(ch chan string) *http.Server {
	mux := http.NewServeMux()
	srv := &http.Server{Addr: ":3000", Handler: mux}
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		if code != "" {
			fmt.Fprintln(w, "Login success. You can now close this window")
			ch <- code
		}
	})
	go srv.ListenAndServe()
	return srv
}
