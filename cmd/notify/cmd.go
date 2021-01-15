package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/tomoyamachi/notifyhome/pkg/cast"
	"github.com/tomoyamachi/notifyhome/pkg/gcal"
	"github.com/urfave/cli/v2"
	calendar "google.golang.org/api/calendar/v3"
)

const (
	notifyInterval = time.Minute * 30
	notifyBuffer   = time.Hour * 3
)

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
			events, err := FetchEvents(cli, 0)
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

func FetchEvents(cli *http.Client, duration time.Duration) ([]*calendar.Event, error) {
	srv, err := calendar.New(cli)
	if err != nil {
		return nil, fmt.Errorf("Retrieve Calendar client: %w", err)
	}
	t := time.Now().Format(time.RFC3339)
	call := srv.Events.List("primary").ShowDeleted(false).
		SingleEvents(true).TimeMin(t).MaxResults(1).OrderBy("startTime")
	if duration > 0 {
		call.TimeMax(time.Now().Add(duration).Format(time.RFC3339))
	}
	events, err := call.Do()
	if err != nil {
		return nil, fmt.Errorf("Retrieve next events: %w", err)
	}
	return events.Items, nil
}

func startDaemon(c *cli.Context) error {
	ticker := time.NewTicker(notifyInterval)
	defer ticker.Stop()
	i := 0
	for {
		select {
		case <-ticker.C:
			i++
			log.Print("start notify plan:", i)
			if err := notifyPlans(c.Context); err != nil {
				log.Print(err)
			}
		}
	}
}

func notifyPlans(ctx context.Context) error {
	clis, err := gcal.GetClients(ctx)
	if err != nil {
		return err
	}
	var wg sync.WaitGroup
	msgCh := make(chan string, len(clis))
	errChan := make(chan error, len(clis))
	i := 0
	for _, cli := range clis {
		i++
		wg.Add(1)
		go func(idx int, cli *http.Client, wg *sync.WaitGroup) {
			defer wg.Done()
			events, err := FetchEvents(cli, notifyBuffer)
			if err != nil {
				errChan <- err
				return
			}
			if len(events) > 0 {
				for _, item := range events {
					date := item.Start.DateTime
					if date == "" {
						date = item.Start.Date
					}
					tm, err := time.Parse(time.RFC3339, date)
					if err != nil {
						errChan <- err
						return
					}
					msgCh <- fmt.Sprintf("%sから%s", tm.Format("15時04分"), item.Summary)
				}
			}
		}(i, cli, &wg)
	}
	wg.Wait()
	close(msgCh)
	close(errChan)

	for msg := range msgCh {
		if err := notifyWithCtx(ctx, "ja", msg); err != nil {
			return err
		}
	}

	for err := range errChan {
		return err
	}
	return nil
}

func notify(c *cli.Context) error {
	return notifyWithCtx(c.Context, "en", "Good Morning. Hello. Good Evening.")
}

func notifyWithCtx(ctx context.Context, locale, msg string) error {
	devices := cast.LookupAndConnect(ctx)
	var wg sync.WaitGroup
	errChan := make(chan error, len(devices))
	for _, device := range devices {
		wg.Add(1)
		go func(ctx context.Context, device *cast.CastDevice, wg *sync.WaitGroup) {
			defer wg.Done()
			if err := device.Speak(ctx, msg, locale); err != nil {
				errChan <- err
			}
		}(ctx, device, &wg)
	}
	wg.Wait()
	close(errChan)
	for err := range errChan {
		return err
	}
	return nil
}
