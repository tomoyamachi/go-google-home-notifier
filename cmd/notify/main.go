package main

import (
	"log"
	"os"

	"github.com/tomoyamachi/notifyhome/pkg/cli"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	if err := cli.App().Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
