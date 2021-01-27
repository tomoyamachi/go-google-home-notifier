package cli

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/urfave/cli/v2"
	"golang.org/x/sync/errgroup"

	"github.com/tomoyamachi/notifyhome/pkg/gcal"
	"github.com/tomoyamachi/notifyhome/pkg/googlecast"
	"github.com/tomoyamachi/notifyhome/pkg/locale"
	"github.com/tomoyamachi/notifyhome/pkg/server"
)

// calendar add-token Action
func addToken(c *cli.Context) error {
	return gcal.AddToken(c.Context)
}

// calendar fetch-plan Action
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

// notify Action
func notifyFromDevices(c *cli.Context) error {
	return googlecast.Notify(c.Context, c.Int("device-count"), c.String("device-name"), c.String("locale"), c.String("message"))
}

// server Action
func simpleServe(c *cli.Context) error {
	deviceCnt := c.Int("device-count")
	deviceName := c.String("device-name")

	return server.Run(c.Context, deviceCnt, deviceName, "ja", c.Int("port"))
}

// daemon Action
func startDaemon(c *cli.Context) error {
	log.Print("Start daemon.")
	ctx, cancel := context.WithCancel(c.Context)
	defer cancel()
	killSignal := make(chan os.Signal, 1)
	signal.Notify(killSignal, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-killSignal
		log.Println("interrupted")
		cancel()
	}()

	eg, ctx := errgroup.WithContext(ctx)
	localeCode := c.String("locale")
	deviceCnt := c.Int("device-count")
	deviceName := c.String("device-name")
	eg.Go(func() error {
		return regularNotify(ctx, deviceCnt, deviceName, localeCode, c.Duration("notify-duration"), c.Duration("within"))
	})
	eg.Go(func() error {
		return server.Run(ctx, deviceCnt, deviceName, localeCode, c.Int("port"))
	})

	return eg.Wait()
}

func regularNotify(ctx context.Context, deviceCnt int, deviceName, localeCode string, tick, within time.Duration) error {
	if err := fetchAndNotifyPlans(ctx, deviceCnt, deviceName, localeCode, within); err != nil {
		return err
	}
	ticker := time.NewTicker(tick)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			log.Print("fetch plans and send notifications")
			if err := fetchAndNotifyPlans(ctx, deviceCnt, deviceName, localeCode, within); err != nil {
				log.Print(err)
			}
		case <-ctx.Done():
			return nil
		}
	}
}

func fetchAndNotifyPlans(ctx context.Context, deviceCnt int, deviceName, localeCode string, within time.Duration) error {
	clis, err := gcal.GetClients(ctx)
	if err != nil {
		return err
	}
	eventsList, errs := getEventsAndEror(clis, 1, within)
	locale := locale.GetLocale(localeCode)
	for _, events := range eventsList {
		for _, event := range events {
			if err := googlecast.Notify(ctx, deviceCnt, deviceName, locale.Code(), locale.NotifyMessage(event.Start, event.Title)); err != nil {
				errs = append(errs, err)
			}
		}
	}
	return checkErrs(errs)
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
