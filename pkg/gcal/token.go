package gcal

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"golang.org/x/oauth2"
)

func AddToken(ctx context.Context, credentialPath string) error {
	config, err := getConfig(credentialPath)
	if err != nil {
		return err
	}

	tok, err := getTokenFromWeb(ctx, config)
	if err != nil {
		return err
	}

	tokens, err := loadTokens(credentialPath)
	if err != nil {
		return err
	}
	tokens = append(tokens, tok)
	return tokens.save(credentialPath)
}

// Request a token from the web, then returns the retrieved token.
func getTokenFromWeb(ctx context.Context, config *oauth2.Config) (*oauth2.Token, error) {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL)
	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		return nil, fmt.Errorf("read authorization code: %w", err)
	}

	tok, err := config.Exchange(ctx, authCode)
	if err != nil {
		return nil, fmt.Errorf("retrieve token from web: %w", err)
	}
	return tok, nil
}

// Load tokens from file.
func loadTokens(credentialPath string) (ts Tokens, err error) {
	f, err := os.Open(credentialPath + tokensFile)
	if err != nil {
		if os.IsNotExist(err) {
			return Tokens{}, nil
		}
		return nil, fmt.Errorf("Open %s: %w", credentialPath+tokensFile, err)
	}
	defer f.Close()
	if err = json.NewDecoder(f).Decode(&ts); err != nil {
		return nil, fmt.Errorf("Decode tokens: %w", err)
	}
	return ts, nil
}

// Saves tokens to a file.
func (t Tokens) save(credentialPath string) error {
	fmt.Printf("Saving credential file to: %s\n", credentialPath+tokensFile)
	f, err := os.OpenFile(credentialPath+tokensFile, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("Save token: %w", err)
	}
	defer f.Close()
	return json.NewEncoder(f).Encode(t)
}
