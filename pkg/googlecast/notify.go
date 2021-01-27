package googlecast

import (
	"context"
	"sync"
)

func Notify(ctx context.Context, deviceCnt int, friendlyName, locale, msg string) error {
	devices := LookupAndConnect(ctx, deviceCnt, friendlyName)
	var wg sync.WaitGroup
	errChan := make(chan error, len(devices))
	for _, device := range devices {
		wg.Add(1)
		go func(ctx context.Context, device *CastDevice, wg *sync.WaitGroup) {
			defer wg.Done()
			if err := device.Speak(ctx, msg, locale); err != nil {
				errChan <- err
			}
		}(ctx, device, &wg)
	}
	wg.Wait()
	close(errChan)
	// TODO: fix: Only return first error
	for err := range errChan {
		return err
	}
	return nil
}
