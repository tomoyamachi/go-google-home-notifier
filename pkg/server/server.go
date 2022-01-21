package server

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/tomoyamachi/notifyhome/pkg/googlecast"
)

func Run(ctx context.Context, deviceCnt int, deviceName, localeCode string, port int) error {
	handler := http.NewServeMux()
	handler.HandleFunc("/notify", func(w http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodPost {
			writeResponse(w, []byte("Invalid methods\n"))
			return
		}
		b, err := ioutil.ReadAll(req.Body)
		if err != nil {
			writeResponse(w, []byte("Internal error\n"))
			return
		}
		if err := googlecast.Notify(ctx, deviceCnt, deviceName, localeCode, []string{string(b)}); err != nil {
			log.Printf("notifyWithCtx %+v\n", err)
			writeResponse(w, []byte("Internal error\n"))
			return
		}
		writeResponse(w, b)
	})
	handler.HandleFunc("/quiet", func(w http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodGet {
			writeResponse(w, []byte("Invalid methods\n"))
			return
		}
		targetTime := time.Now().Add(24 * time.Hour)
		googlecast.SetNotifyAfter(targetTime)
		writeResponse(w, []byte(fmt.Sprintf("I will be quiet until %s\n", targetTime.Format("2006/01/02 15:04"))))
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

func writeResponse(w http.ResponseWriter, b []byte) {
	if _, err := w.Write(b); err != nil {
		log.Printf("write to body %+v\n", err)
	}
}
