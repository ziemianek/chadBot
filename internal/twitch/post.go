package twitch

import (
	"bytes"
	"encoding/json"
	"net/http"

	_ "github.com/joho/godotenv/autoload"
)

// structs below are used to send POST request
type content struct {
	Type      string    `json:"type"`
	Version   string    `json:"version"`
	Condition condition `json:"condition"`
	Transport transport `json:"transport"`
}

type condition struct {
	BroadcasterUserId string `json:"broadcaster_user_id"`
	UserId            string `json:"user_id"`
}

type transport struct {
	Method    string `json:"method"`
	SessionId string `json:"session_id"`
}

func SendPOST(url string, content content, headers map[string]string) (*http.Response, error) {
	var err error
	var body []byte
	var request *http.Request
	body, err = json.Marshal(content)
	if err != nil {
		return nil, err
	}
	request, err = http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	for k, v := range headers {
		request.Header.Set(k, v)
	}
	return http.DefaultClient.Do(request)
}
