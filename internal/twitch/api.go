package twitch

import (
	"bytes"
	"errors"
	"net/http"
	"os"
)

type headers map[string]string

func callAPI(method, url string, headers headers, body []byte) (*http.Response, error) {
	if !(method == http.MethodGet || method == http.MethodPost) {
		return &http.Response{},
			errors.New("Invalid HTTP method. Use http.MethodGet or http.MethodPost")
	}

	var err error
	var req *http.Request
	req, err = http.NewRequest(method, url, bytes.NewReader(body))
	if err != nil {
		return &http.Response{}, err
	}

	req.Header.Set("Authorization", "Bearer "+os.Getenv("ACCESS_TOKEN"))
	req.Header.Set("Client-Id", os.Getenv("CLIENT_ID"))
	// add any additional headers
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	return http.DefaultClient.Do(req)
}

func SendGet(url string) (*http.Response, error) {
	return callAPI(http.MethodGet, url, nil, nil)
}

func SendPost(url string, headers headers, body []byte) (*http.Response, error) {
	return callAPI(http.MethodPost, url, headers, body)
}
