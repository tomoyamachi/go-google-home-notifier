package gcal

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
)

const (
	tokensFile     = "tokens.json"
	credentialFile = "credentials.json"
)

type Tokens []*oauth2.Token

func getConfig() (*oauth2.Config, error) {
	b, err := ioutil.ReadFile(credentialFile)
	if err != nil {
		return nil, fmt.Errorf("Read google client secret: %w", err)
	}
	config, err := google.ConfigFromJSON(b, calendar.CalendarReadonlyScope)
	if err != nil {
		return nil, fmt.Errorf("Load config from secret: %w", err)
	}
	return config, nil
}

// Retrieve a token, saves the token, then returns the generated client.
func GetClients(ctx context.Context) ([]*http.Client, error) {
	config, err := getConfig()
	if err != nil {
		return nil, err
	}

	tokens, err := loadTokens()
	if err != nil {
		return nil, err
	}

	clis := make([]*http.Client, len(tokens))
	for idx, tok := range tokens {
		clis[idx] = config.Client(ctx, tok)
	}
	return clis, nil
}
