package googlecast

import (
	"context"
	"log"
	"time"
)

var notifyAfter = time.Now()

func SetNotifyAfter(target time.Time) {
	notifyAfter = target
}

func notifiable() bool {
	return notifyAfter.Before(time.Now())
}

func Notify(ctx context.Context, deviceCnt int, friendlyName, locale string, msgs []string) error {
	if !notifiable() {
		log.Printf("notify will restart after %s", notifyAfter.Format("2006/01/02 15:04"))
		return nil
	}

	if len(msgs) == 0 {
		return nil
	}
	devices := LookupAndConnect(ctx, deviceCnt, friendlyName)
	if len(devices) == 0 {
		log.Print("no device found.")
		return nil
	}
	errs := []error{}
	for _, device := range devices {
		totalMsg := ""
		for _, msg := range msgs {
			totalMsg += msg
		}

		if len(totalMsg) > 0 {
			if err := device.Speak(ctx, totalMsg, locale); err != nil {
				errs = append(errs, err)
			}
		}
	}
	// TODO: fix: Only return first error
	for _, err := range errs {
		return err
	}
	return nil
}
