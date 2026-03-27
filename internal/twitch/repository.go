package twitch

import "context"

type SecretsRepository interface {
	GetCredentials() (clientID, clientSecret string, err error)
	GetToken(ctx context.Context) (*TwitchToken, error)
	SaveToken(ctx context.Context, token *TwitchToken) error
}
