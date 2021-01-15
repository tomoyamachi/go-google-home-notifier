package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"sort"
	"sync"
	"time"

	"github.com/urfave/cli/v2"
	"google.golang.org/api/calendar/v3"

	"github.com/tomoyamachi/notifyhome/pkg/cast"
	"github.com/tomoyamachi/notifyhome/pkg/gcal"
)

func main() {
	app := &cli.App{
		Commands: []*cli.Command{
			{
				Name:    "daemon",
				Aliases: []string{"d"},
				Usage:   "start notify daemon",
				Action:  startDaemon,
			},
			{
				Name:    "calendar",
				Aliases: []string{"c"},
				Usage:   "about google calendar",
				Subcommands: []*cli.Command{
					{
						Name:    "add-token",
						Aliases: []string{"a"},
						Usage:   "add google account",
						Action:  addToken,
					},
					{
						Name:    "fetch-plan",
						Aliases: []string{"f"},
						Usage:   "fetch google calendar",
						Action:  fetchPlan,
					},
				},
			},
			{
				Name:  "notify",
				Usage: "simple notify",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "message",
						Aliases: []string{"m"},
					},
				},
				Action: notify,
			},
		},
	}

	sort.Sort(cli.FlagsByName(app.Flags))
	sort.Sort(cli.CommandsByName(app.Commands))

	err := app.Run(os.Args)

	if err != nil {
		log.Fatal(err)
	}
}

func addToken(c *cli.Context) error {
	return gcal.AddToken(c.Context)
}

func fetchPlan(c *cli.Context) error {
	clis, err := gcal.GetClients(c.Context)
	if err != nil {
		return err
	}
	var wg sync.WaitGroup
	errChan := make(chan error, len(clis))
	i := 0
	for _, cli := range clis {
		i++
		wg.Add(1)

		go func(idx int, cli *http.Client, wg *sync.WaitGroup) {
			defer wg.Done()
			events, err := FetchEvents(cli)
			if err != nil {
				errChan <- err
			}
			if len(events) == 0 {
				fmt.Printf("Account %d: No upcoming events\n", idx)
			} else {
				for _, item := range events {
					date := item.Start.DateTime
					if date == "" {
						date = item.Start.Date
					}
					fmt.Printf("Account %d: %v %v\n", idx, date, item.Summary)
				}
			}
		}(i, cli, &wg)
	}
	wg.Wait()
	close(errChan)
	for err := range errChan {
		return err
	}
	return nil
}

func FetchEvents(cli *http.Client) ([]*calendar.Event, error) {
	srv, err := calendar.New(cli)
	if err != nil {
		return nil, fmt.Errorf("Retrieve Calendar client: %w", err)
	}
	t := time.Now().Format(time.RFC3339)
	events, err := srv.Events.List("primary").ShowDeleted(false).
		SingleEvents(true).TimeMin(t).MaxResults(10).OrderBy("startTime").Do()
	if err != nil {
		return nil, fmt.Errorf("Retrieve next events: %w", err)
	}
	return events.Items, nil
}

func startDaemon(c *cli.Context) error {
	return gcal.AddToken(c.Context)
}

func notify(c *cli.Context) error {
	devices := cast.LookupAndConnect(c.Context)
	var wg sync.WaitGroup
	errChan := make(chan error, len(devices))
	for _, device := range devices {
		wg.Add(1)
		go func(ctx context.Context, device *cast.CastDevice, wg *sync.WaitGroup) {
			defer wg.Done()
			if err := device.Speak(ctx, "Hello World", "en"); err != nil {
				errChan <- err
			}
		}(c.Context, device, &wg)
	}
	wg.Wait()
	close(errChan)
	for err := range errChan {
		return err
	}
	return nil
}
