package twitch

// TODO: OMG REFACTOR THIS SPAGHETTI GIANT

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	_ "github.com/joho/godotenv/autoload"

	"github.com/charmbracelet/log"
	"github.com/gorilla/websocket"
	"github.com/toqueteos/webbrowser"
)

type TwitchToken struct {
	AccessToken  string   `json:"access_token"`
	RefreshToken string   `json:"refresh_token"`
	ExpiresIn    int      `json:"expires_in"`
	Scope        []string `json:"scope"`
}

type TwitchUserResponse struct {
	Data []struct {
		ID              string `json:"id"`
		Login           string `json:"login"`
		DisplayName     string `json:"display_name"`
		Description     string `json:"description"`
		ProfileImageURL string `json:"profile_image_url"`
	} `json:"data"`
}

type Client struct {
	conn         *websocket.Conn
	state        string
	accessToken  *TwitchToken
	clientId     string
	clientSecret string
}

func NewClient() *Client {
	var state string
	var err error
	state, err = generateState()
	if err != nil {
		log.Errorf("Could not generate state: %v", err)
		return &Client{}
	}

	var clientId string
	var clientSecret string
	clientId, clientSecret, err = getCredentials()
	if err != nil {
		log.Errorf("Could not get credentials: %v", err)
		return &Client{}
	}
	return &Client{
		conn:         &websocket.Conn{},
		state:        state,
		accessToken:  &TwitchToken{},
		clientId:     clientId,
		clientSecret: clientSecret,
	}
}

func getCredentials() (string, string, error) {
	var exists bool
	var clientId string
	clientId, exists = os.LookupEnv("CLIENT_ID")
	if !exists {
		return "", "", errors.New("CLIENT_ID not set")
	}
	var clientSecret string
	clientSecret, exists = os.LookupEnv("CLIENT_SECRET")
	if !exists {
		log.Errorf("CLIENT_SECRET not set")
		return "", "", errors.New("CLIENT_SECRET not set")
	}
	return clientId, clientSecret, nil
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
	var resChan chan string = make(chan string)
	var authUrl string = c.generateAuthURL()
	startServer(resChan)
	if err := webbrowser.Open(authUrl); err != nil {
		return err
	}
	select {
	case code := <-resChan:
		accessToken, err := c.generateToken(code)
		if err != nil {
			log.Errorf("Could not generate access token: %v", err)
			return errors.New("Error generating access token")
		}
		c.accessToken = accessToken
		return nil
	case <-time.After(2 * time.Minute):
		return errors.New("timed out")
	}
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
		err = welcomeMessage.handle(c, msg)
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

func (c Client) generateToken(code string) (*TwitchToken, error) {
	data := url.Values{}
	data.Set("client_id", c.clientId)
	data.Set("client_secret", c.clientSecret)
	data.Set("code", code)
	data.Set("grant_type", "authorization_code")
	data.Set("redirect_uri", "http://localhost:3000")

	resp, err := http.Post("https://id.twitch.tv/oauth2/token", "application/x-www-form-urlencoded", strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var t TwitchToken
	json.NewDecoder(resp.Body).Decode(&t)
	return &t, nil
}

type RequestHeaders map[string]string

func (c *Client) callAPI(method, url string, headers RequestHeaders, body []byte) (*http.Response, error) {
	if !(method == http.MethodGet || method == http.MethodPost) {
		//TODO: use nil???
		return &http.Response{},
			errors.New("Invalid HTTP method. Use http.MethodGet or http.MethodPost")
	}

	var err error
	var req *http.Request
	req, err = http.NewRequest(method, url, bytes.NewReader(body))
	if err != nil {
		return &http.Response{}, err
	}

	if c.accessToken != nil {
		req.Header.Set("Authorization", "Bearer "+c.accessToken.AccessToken)
		req.Header.Set("Client-Id", c.clientId)
		// add any additional headers
		for k, v := range headers {
			req.Header.Set(k, v)
		}

		return http.DefaultClient.Do(req)
	} else {
		return nil, errors.New("No valid access token in configuration")
	}
}

func (c Client) SendGet(url string) (*http.Response, error) {
	return c.callAPI(http.MethodGet, url, nil, nil)
}

func (c Client) SendPost(url string, headers RequestHeaders, body []byte) (*http.Response, error) {
	return c.callAPI(http.MethodPost, url, headers, body)
}

type payloadSubscribeToChat struct {
	Type      string `json:"type"`
	Version   string `json:"version"`
	Condition struct {
		BroadcasterUserId string `json:"broadcaster_user_id"`
		UserId            string `json:"user_id"`
	} `json:"condition"`
	Transport struct {
		Method    string `json:"method"`
		SessionId string `json:"session_id"`
	} `json:"transport"`
}

type welcomeMessage struct {
	Payload struct {
		Session struct {
			Id                  string `json:"id"`
			Status              string `json:"status"`
			ConnectedAt         string `json:"connected_at"`
			KeepaliveTimeoutSec int    `json:"keepalive_timeout_seconds"`
			ReconnectUrl        string `json:"reconnect_url"`
			RecoveryUrl         string `json:"recovery_url"`
		} `json:"session"`
	} `json:"payload"`
}

func (m *welcomeMessage) handle(client Client, msg []byte) error {
	var err error
	var sessionId string
	err = json.Unmarshal(msg, m)
	sessionId = m.Payload.Session.Id
	log.Infof("Extracted session id: %v", sessionId)
	client.subscribe(sessionId)
	return err
}

type notificationMessage struct {
	Metadata struct {
		MsgTimestamp string `json:"message_timestamp"`
	} `json:"metadata"`
	Payload struct {
		Event struct {
			ChatterUserName string `json:"chatter_user_name"`
			Message         struct {
				Text string `json:"text"`
			} `json:"message"`
		} `json:"event"`
	} `json:"payload"`
}

func (m *notificationMessage) handle(ch chan string, msg []byte) error {
	var err error
	err = json.Unmarshal(msg, m)
	log.Debugf("Parsing notification message: %v", string(msg))
	// TODO: refactor this maybe?
	ch <- fmt.Sprintf("[%v] - [%v]: %v",
		parseTimestamp(m.Metadata.MsgTimestamp),
		m.Payload.Event.ChatterUserName,
		m.Payload.Event.Message.Text,
	)
	log.Infof("Sent new message: \"%v\" to message channel", m.Payload.Event.Message.Text)
	return err
}

func (c Client) subscribe(sessionId string) error {
	var url string = "https://api.twitch.tv/helix/eventsub/subscriptions"
	var err error
	var resp *http.Response
	//TODO: if all post requests get this, then move to callAPI method
	var headers RequestHeaders = RequestHeaders{"Content-Type": "application/json"}
	var broadcasterID string = c.getBroadcasterID()
	var body []byte
	body, err = json.Marshal(payloadSubscribeToChat{
		Type:    "channel.chat.message",
		Version: "1",
		Condition: struct {
			BroadcasterUserId string `json:"broadcaster_user_id"`
			UserId            string `json:"user_id"`
		}{
			BroadcasterUserId: broadcasterID,
			UserId:            broadcasterID,
		},
		Transport: struct {
			Method    string `json:"method"`
			SessionId string `json:"session_id"`
		}{
			Method:    "websocket",
			SessionId: sessionId,
		},
	})
	resp, err = c.SendPost(url, headers, body)
	if err != nil {
		return err
	}
	if resp.StatusCode == http.StatusAccepted {
		log.Info("Successfully subscribed to chat")
	} else {
		// TODO: make this more efficient by streaming data?
		out, _ := io.ReadAll(resp.Body)
		log.Errorf("Could not authorize: %v", string(out))
	}
	return nil
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

func (c Client) getBroadcasterID() string {
	var url string = "https://api.twitch.tv/helix/users"
	var err error
	var resp *http.Response
	resp, err = c.SendGet(url)
	if err != nil {
		log.Errorf("Could not send request: %v", err)
		return ""
	}
	var respBody struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	// stream data directly to struct
	// more efficient than io.ReadAll + json.Unmarshal
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	if err != nil {
		log.Errorf("Could not unmarshal response: %v", err)
		return ""
	}
	if len(respBody.Data) == 0 {
		log.Errorf("Twitch returned 0 users for this token. Check your credentials.")
		return ""
	}
	var id string = respBody.Data[0].ID
	log.Infof("Extracted Broadcaster User ID: %s", id)
	return id
}
