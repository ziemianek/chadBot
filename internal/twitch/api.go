package twitch

import (
	"bytes"
	"errors"
	"net/http"
	"os"

	_ "github.com/joho/godotenv/autoload"
)

type RequestHeaders map[string]string

// TODO: make this twitch Client receiver
func callAPI(method, url string, headers RequestHeaders, body []byte) (*http.Response, error) {
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

	var accessToken string
	var exists bool
	accessToken, exists = os.LookupEnv("ACCESS_TOKEN")
	if !exists {
		//TODO: use nil???
		return &http.Response{},
			errors.New("Env var ACCESS_TOKEN is missing")
	}
	var clientId string
	clientId, exists = os.LookupEnv("CLIENT_ID")
	if !exists {
		//TODO: use nil???
		return &http.Response{},
			errors.New("Env var CLIENT_ID is missing")

	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Client-Id", clientId)
	// add any additional headers
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	return http.DefaultClient.Do(req)
}

func SendGet(url string) (*http.Response, error) {
	return callAPI(http.MethodGet, url, nil, nil)
}

func SendPost(url string, headers RequestHeaders, body []byte) (*http.Response, error) {
	return callAPI(http.MethodPost, url, headers, body)
}
