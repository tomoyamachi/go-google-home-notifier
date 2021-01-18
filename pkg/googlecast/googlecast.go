package googlecast

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"strings"

	cast "github.com/barnybug/go-cast"
	"github.com/barnybug/go-cast/controllers"
	castnet "github.com/barnybug/go-cast/net"
	"github.com/micro/mdns"
)

const (
	googleCastServiceName = "_googlecast._tcp"
	modelTypePrefix       = "md"
	friendryNamePrefix    = "fn"
	googleHomeModelPrefix = "md=Google"
)

// CastDevice is cast-able device contains cast client
type CastDevice struct {
	*mdns.ServiceEntry
	client *cast.Client
}

// Connect connects required services to cast
func (g *CastDevice) Connect(ctx context.Context) error {
	return g.client.Connect(ctx)
}

// Close calls client's close func
func (g *CastDevice) Close() {
	g.client.Close()
}

// Speak speaks given text on cast device
func (g *CastDevice) Speak(ctx context.Context, text, lang string) error {
	url, err := tts(text, lang)
	if err != nil {
		return err
	}
	return g.Play(ctx, url)
}

// LookupAndConnect retrieves cast-able google home devices
func LookupAndConnect(ctx context.Context, max int, friendryName string) []*CastDevice {
	// https://github.com/hashicorp/mdns
	entriesCh := make(chan *mdns.ServiceEntry, max)
	resultCh := make(chan *CastDevice, max)
	go func(ctx context.Context, friendryName string) {
		for entry := range entriesCh {
			if cast := lookupClient(ctx, entry, friendryName); cast != nil {
				resultCh <- cast
			}
		}
	}(ctx, friendryName)
	mdns.Lookup(googleCastServiceName, entriesCh)
	close(entriesCh)
	close(resultCh)
	results := make([]*CastDevice, 0, max)
	for cast := range resultCh {
		if cast != nil {
			results = append(results, cast)
		}
	}
	return results
}

func lookupClient(ctx context.Context, entry *mdns.ServiceEntry, friendryName string) *CastDevice {
	var client *cast.Client
	valid := true
	// Fields : https://blog.oakbits.com/google-cast-protocol-discovery-and-connection.html
	for _, field := range entry.InfoFields {
		// check device friendly name
		if friendryName != "" && strings.HasPrefix(field, friendryNamePrefix) {
			if field != fmt.Sprintf("%s=%s", friendryNamePrefix, friendryName) {
				valid = false
				continue
			}
		}
		if strings.HasPrefix(field, modelTypePrefix) {
			if !strings.HasPrefix(field, googleHomeModelPrefix) {
				valid = false
				continue
			}
			client = cast.NewClient(entry.AddrV4, entry.Port)
			if err := client.Connect(ctx); err != nil {
				log.Printf("[ERROR] Failed to connect: %s", err)
			}
		}
	}
	if valid {
		return &CastDevice{entry, client}
	}
	return nil
}

// tts provides text-to-speech sound url.
// NOTE: it seems to be unofficial.
func tts(text, lang string) (*url.URL, error) {
	base := "https://translate.google.com/translate_tts?client=tw-ob&ie=UTF-8&q=%s&tl=%s"
	return url.Parse(fmt.Sprintf(base, url.QueryEscape(text), url.QueryEscape(lang)))
}

// Play plays media contents on cast device
func (g *CastDevice) Play(ctx context.Context, url *url.URL) error {
	conn := castnet.NewConnection()
	if err := conn.Connect(ctx, g.AddrV4, g.Port); err != nil {
		return err
	}
	defer conn.Close()
	status, err := g.client.Receiver().LaunchApp(ctx, cast.AppMedia)
	if err != nil {
		return err
	}
	app := status.GetSessionByAppId(cast.AppMedia)
	cc := controllers.NewConnectionController(conn, g.client.Events, cast.DefaultSender, *app.TransportId)
	if err := cc.Start(ctx); err != nil {
		return err
	}
	media := controllers.NewMediaController(conn, g.client.Events, cast.DefaultSender, *app.TransportId)
	if err := media.Start(ctx); err != nil {
		return err
	}

	mediaItem := controllers.MediaItem{
		ContentId:   url.String(),
		ContentType: "audio/mp3",
		StreamType:  "BUFFERED",
	}

	log.Printf("[INFO] Load media: content_id=%s", mediaItem.ContentId)
	_, err = media.LoadMedia(ctx, mediaItem, 0, true, nil)

	return err
}
