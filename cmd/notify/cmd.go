package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/urfave/cli/v2"

	"github.com/tomoyamachi/notifyhome/pkg/cast"
	"github.com/tomoyamachi/notifyhome/pkg/gcal"
	"github.com/tomoyamachi/notifyhome/pkg/locale"
)

func addToken(c *cli.Context) error {
	return gcal.AddToken(c.Context)
}

func notifyFromDevices(c *cli.Context) error {
	return notifyWithCtx(c.Context, c.String("locale"), c.String("message"))
}

func fetchAndShowPlans(c *cli.Context) error {
	clis, err := gcal.GetClients(c.Context)
	if err != nil {
		return err
	}
	eventsList, errs := getEventsAndEror(clis, c.Int64("count"), c.Duration("within"))
	for idx, events := range eventsList {
		for _, event := range events {
			fmt.Printf("%d: %v %s\n", idx, event.Start, event.Title)
		}
	}
	return checkErrs(errs)
}

func startDaemon(c *cli.Context) error {
	log.Print("Start daemon.")
	within := c.Duration("within")
	localeCode := c.String("locale")
	if err := fetchAndNotifyPlans(c.Context, localeCode, within); err != nil {
		return err
	}
	ticker := time.NewTicker(c.Duration("notify-duration"))
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			log.Print("run fetchAndNotifyPlans")
			if err := fetchAndNotifyPlans(c.Context, localeCode, within); err != nil {
				log.Print(err)
			}
		}
	}
}

func fetchAndNotifyPlans(ctx context.Context, localeCode string, within time.Duration) error {
	clis, err := gcal.GetClients(ctx)
	if err != nil {
		return err
	}
	eventsList, errs := getEventsAndEror(clis, 1, within)
	locale := locale.GetLocale(localeCode)
	for _, events := range eventsList {
		for _, event := range events {
			if err := notifyWithCtx(ctx, locale.Code(), locale.NotifyMessage(event.Start, event.Title)); err != nil {
				errs = append(errs, err)
			}
		}
	}
	return checkErrs(errs)
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

func getEventsAndEror(clis []*http.Client, cnt int64, within time.Duration) ([][]*gcal.Event, []error) {
	eventsCh := make(chan []*gcal.Event, len(clis))
	errChan := make(chan error, len(clis))
	var wg sync.WaitGroup
	wg.Add(len(clis))
	for _, cli := range clis {
		go func(cli *http.Client, wg *sync.WaitGroup) {
			defer wg.Done()
			events, err := gcal.FetchEvents(cli, cnt, within)
			if err != nil {
				errChan <- err
				return
			}
			if len(events) > 0 {
				eventsCh <- events
			}
		}(cli, &wg)
	}
	wg.Wait()
	close(eventsCh)
	close(errChan)

	eventsList := make([][]*gcal.Event, len(clis))
	for event := range eventsCh {
		eventsList = append(eventsList, event)
	}

	errs := []error{}
	for err := range errChan {
		errs = append(errs, err)
	}
	return eventsList, errs
}

func checkErrs(errs []error) (err error) {
	if len(errs) == 0 {
		return nil
	}
	for _, err = range errs {
		log.Print(err)
	}
	return err // temporary, return a last error
}
