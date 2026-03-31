package twitch

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/charmbracelet/log"
)

// TwitchToken holds the OAuth2 credentials.
type TwitchToken struct {
	AccessToken  string   `json:"access_token"`
	RefreshToken string   `json:"refresh_token"`
	ExpiresIn    int      `json:"expires_in"`
	Scope        []string `json:"scope"`

	CreatedAt time.Time
	ExpiresAt time.Time
	Expired   bool
}

func NewTwitchToken(r *http.Response) (*TwitchToken, error) {
	var t *TwitchToken
	if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
		return nil, err
	}
	t.CreatedAt = time.Now()
	t.ExpiresAt = t.CreatedAt.Add(time.Duration(t.ExpiresIn) * time.Second)
	return t, nil
}

func (t *TwitchToken) isExpired() bool {
	if t.Expired {
		return true
	}
	if t.AccessToken == "" {
		log.Error("Couldnt validate access token, because AccessToken is empty string")
		return true
	}
	// checks if time is actually recordered
	if !t.CreatedAt.IsZero() {
		t.Expired = time.Now().After(t.ExpiresAt.Add(-60 * time.Second))
		return t.Expired
	}
	log.Errorf("Couldnt validate access token, most likely CreatedAt is not set")
	return true
}

func (t *TwitchToken) Refresh(ctx context.Context, id, secret string) (*TwitchToken, error) {
	//FIXME: duplicate code of client.generateToken
	data := url.Values{}
	data.Set("client_id", id)
	data.Set("client_secret", secret)
	data.Set("grant_type", "refresh_token")
	data.Set("refresh_token", t.RefreshToken)

	// returns HTTP 400 if the request fails
	// source: https://dev.twitch.tv/docs/authentication/refresh-tokens/
	resp, err := http.Post(TwitchTokenURL, "application/x-www-form-urlencoded", strings.NewReader(data.Encode()))
	if err != nil {
		return t, err
	}
	defer resp.Body.Close()

	var respBody struct {
		AccessToken  string   `json:"access_token"`
		RefreshToken string   `json:"refresh_token"`
		Scope        []string `json:"scope"`
		TokenType    string   `json:"token_type"`
	}
	// stream data directly to struct
	// more efficient than io.ReadAll + json.Unmarshal
	if err = json.NewDecoder(resp.Body).Decode(&respBody); err != nil {
		log.Errorf("Could not unmarshal response: %v", err)
		return t, err
	}
	log.Debugf("Refresh token response: %v", respBody)
	t.AccessToken = respBody.AccessToken
	t.RefreshToken = respBody.RefreshToken
	t.CreatedAt = time.Now()
	// those do not come in a response from token refresh
	// i need to think of a way to get rid of them for one
	t.ExpiresIn = 3 //hours
	t.ExpiresAt = t.CreatedAt.Add(time.Duration(t.ExpiresIn) * time.Hour)
	t.Expired = false
	return t, nil
}
