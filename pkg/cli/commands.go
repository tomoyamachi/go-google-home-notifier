package cli

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
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
)

func addToken(c *cli.Context) error {
	return gcal.AddToken(c.Context)
}

func notifyFromDevices(c *cli.Context) error {
	return notifyWithCtx(c.Context, c.Int("device-count"), c.String("device-name"), c.String("locale"), c.String("message"))
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
	eg, ctx := errgroup.WithContext(c.Context)
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	killSignal := make(chan os.Signal, 1)
	signal.Notify(killSignal, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-killSignal
		log.Println("interrupted")
		cancel()
	}()

	localeCode := c.String("locale")
	deviceCnt := c.Int("device-count")
	deviceName := c.String("device-name")
	eg.Go(func() error {
		return regularNotify(ctx, deviceCnt, deviceName, localeCode, c.Duration("notify-duration"), c.Duration("within"))
	})
	eg.Go(func() error {
		return httpRun(ctx, deviceCnt, deviceName, localeCode, c.Int("port"))
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
			if err := notifyWithCtx(ctx, deviceCnt, deviceName, locale.Code(), locale.NotifyMessage(event.Start, event.Title)); err != nil {
				errs = append(errs, err)
			}
		}
	}
	return checkErrs(errs)
}

func notifyWithCtx(ctx context.Context, deviceCnt int, friendlyName, locale, msg string) error {
	devices := googlecast.LookupAndConnect(ctx, deviceCnt, friendlyName)
	var wg sync.WaitGroup
	errChan := make(chan error, len(devices))
	for _, device := range devices {
		wg.Add(1)
		go func(ctx context.Context, device *googlecast.CastDevice, wg *sync.WaitGroup) {
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

func simpleServe(c *cli.Context) error {
	deviceCnt := c.Int("device-count")
	deviceName := c.String("device-name")

	return httpRun(c.Context, deviceCnt, deviceName, "ja", c.Int("port"))
}

func httpRun(ctx context.Context, deviceCnt int, deviceName, localeCode string, port int) error {
	handler := http.NewServeMux()
	handler.HandleFunc("/notify", func(w http.ResponseWriter, req *http.Request) {
		switch req.Method {
		case http.MethodPost:
			b, err := ioutil.ReadAll(req.Body)
			if err != nil {
				if _, err := io.WriteString(w, "Internal error\n"); err != nil {
					log.Printf("write to body %+v\n", err)
				}
				return
			}
			if err := notifyWithCtx(ctx, deviceCnt, deviceName, localeCode, string(b)); err != nil {
				log.Printf("notifyWithCtx %+v\n", err)
				if _, err := io.WriteString(w, "Internal error\n"); err != nil {
					log.Printf("write to body %+v\n", err)
				}
				return
			}
			if _, err := w.Write(b); err != nil {
				log.Printf("write to body %+v\n", err)
			}
			return
		default:
			if _, err := io.WriteString(w, "Invalid methods\n"); err != nil {
				log.Printf("write to body %+v\n", err)
			}
			return
		}
	})
	server := &http.Server{Addr: fmt.Sprintf(":%d", port), Handler: handler}
	go func() {
		<-ctx.Done()
		log.Print("httpRun will be stop...")
		if err := server.Shutdown(ctx); err != nil {
			log.Printf("http server shutdown: %+v\n", err)
		}
	}()
	log.Printf("server start on port: %d\n", port)
	if err := server.ListenAndServe(); err != nil {
		return err
	}
	return nil
}
