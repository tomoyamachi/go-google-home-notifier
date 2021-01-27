package server

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/tomoyamachi/notifyhome/pkg/googlecast"
)

func Run(ctx context.Context, deviceCnt int, deviceName, localeCode string, port int) error {
	handler := http.NewServeMux()
	handler.HandleFunc("/notify", func(w http.ResponseWriter, req *http.Request) {
		switch req.Method {
		case http.MethodPost:
			b, err := ioutil.ReadAll(req.Body)
			if err != nil {
				writeResponse(w, []byte("Internal error\n"))
				return
			}
			if err := googlecast.Notify(ctx, deviceCnt, deviceName, localeCode, string(b)); err != nil {
				log.Printf("notifyWithCtx %+v\n", err)
				writeResponse(w, []byte("Internal error\n"))
				return
			}
			writeResponse(w, b)
		default:
			writeResponse(w, []byte("Invalid methods\n"))
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

func writeResponse(w http.ResponseWriter, b []byte) {
	if _, err := w.Write(b); err != nil {
		log.Printf("write to body %+v\n", err)
	}
}
