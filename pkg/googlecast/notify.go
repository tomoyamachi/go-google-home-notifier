package googlecast

import (
	"context"
	"log"
	"time"
)

func Notify(ctx context.Context, deviceCnt int, friendlyName, locale string, msgs []string) error {
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
		for idx, msg := range msgs {
			if idx != 0 {
				time.Sleep(time.Second * 10)
			}
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
