package twitch

import (
	"encoding/json"
	"net/http"
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

func (t *TwitchToken) refresh() error {
	return nil
}
