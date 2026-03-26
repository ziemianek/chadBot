package twitch

import (
	"bytes"
	"encoding/json"
	"net/http"

	_ "github.com/joho/godotenv/autoload"
)

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
