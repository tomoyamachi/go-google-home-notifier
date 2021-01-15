package main

import (
	"context"
	"log"

	"github.com/tomoyamachi/notifyhome/homecast"
)

func main() {
	ctx := context.Background()
	devices := homecast.LookupAndConnect(ctx)
	for _, device := range devices {
		if err := device.Speak(ctx, "Hello World", "en"); err != nil {
			log.Printf("device %s : %+v", device.Name, err)
		}
	}
}