package gcal

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
)

const (
	tokensFile     = "tokens.json"
	credentialFile = "credentials.json"
)

type (
	Tokens []*oauth2.Token
	Event  struct {
		Title string
		Start time.Time
	}
)

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

func FetchEvents(cli *http.Client, max int64, duration time.Duration) ([]*Event, error) {
	events, err := fetchFromGoogle(cli, max, duration)
	if err != nil {
		return nil, err
	}
	es := make([]*Event, len(events))
	for idx, item := range events {
		var e *Event
		if e, err = parseEvent(item); err != nil {
			return nil, err
		}
		es[idx] = e
	}
	return es, nil
}

func fetchFromGoogle(cli *http.Client, max int64, duration time.Duration) ([]*calendar.Event, error) {
	srv, err := calendar.New(cli)
	if err != nil {
		return nil, fmt.Errorf("Retrieve client: %w", err)
	}
	t := time.Now().Format(time.RFC3339)
	call := srv.Events.List("primary").ShowDeleted(false).
		SingleEvents(true).TimeMin(t).MaxResults(max).OrderBy("startTime")
	if duration > 0 {
		call.TimeMax(time.Now().Add(duration).Format(time.RFC3339))
	}
	events, err := call.Do()
	if err != nil {
		return nil, fmt.Errorf("Retrieve next events: %w", err)
	}
	return events.Items, nil
}

func parseEvent(item *calendar.Event) (*Event, error) {
	date := item.Start.DateTime
	if date == "" {
		date = item.Start.Date
	}
	tm, err := time.Parse(time.RFC3339, date)
	if err != nil {
		return nil, fmt.Errorf("parse event time: %w", err)
	}
	return &Event{Start: tm, Title: item.Summary}, nil
}
