package twitch

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	_ "github.com/joho/godotenv/autoload"
)

type FileSecretsRepository struct {
	tokenPath string
}

func NewFileSecretRepo(path string) *FileSecretsRepository {
	return &FileSecretsRepository{tokenPath: path}
}

func (r *FileSecretsRepository) GetCredentials() (string, string, error) {
	id, ok := os.LookupEnv("CLIENT_ID")
	if !ok {
		return "", "", fmt.Errorf("CLIENT_ID not found in environment")
	}
	secret, ok := os.LookupEnv("CLIENT_SECRET")
	if !ok {
		return "", "", fmt.Errorf("CLIENT_SECRET not found in environment")
	}
	return id, secret, nil
}

func (r *FileSecretsRepository) GetToken(ctx context.Context) (*TwitchToken, error) {
	data, err := os.ReadFile(r.tokenPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, os.ErrNotExist
		}
		return nil, err
	}
	var t TwitchToken
	err = json.Unmarshal(data, &t)
	return &t, err
}

func (r *FileSecretsRepository) SaveToken(ctx context.Context, t *TwitchToken) error {
	data, err := json.MarshalIndent(t, "", " ")
	if err != nil {
		return err
	}
	return os.WriteFile(r.tokenPath, data, 0600)
}
