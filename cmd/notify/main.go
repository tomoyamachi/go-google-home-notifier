package main

import (
	"log"
	"os"
	"sort"
	"time"

	"github.com/urfave/cli/v2"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	app := &cli.App{
		Commands: []*cli.Command{
			{
				Name:    "daemon",
				Aliases: []string{"d"},
				Usage:   "start notifyFromDevices daemon",
				Flags: []cli.Flag{
					&cli.DurationFlag{
						Name:    "notify-duration",
						Aliases: []string{"n"},
						Value:   time.Minute * 30,
						Usage:   "interval between fetch plans and notifyFromDevices",
					},
					&cli.DurationFlag{
						Name:    "within",
						Aliases: []string{"w"},
						Value:   time.Hour * 3,
						Usage:   "fetch plans within target duration from google calendar",
					},
					&cli.StringFlag{
						Name:    "locale",
						Aliases: []string{"l"},
						Value:   "ja",
						Usage:   "message locale code",
					},
				},
				Action: startDaemon,
			},
			{
				Name:    "calendar",
				Aliases: []string{"c"},
				Usage:   "about google calendar",
				Subcommands: []*cli.Command{
					{
						Name:    "add-token",
						Aliases: []string{"a"},
						Usage:   "add google account",
						Action:  addToken,
					},
					{
						Name:    "fetch-plan",
						Aliases: []string{"f"},
						Usage:   "fetch google calendar",
						Action:  fetchAndShowPlans,
						Flags: []cli.Flag{
							&cli.Int64Flag{
								Name:    "count",
								Aliases: []string{"c"},
								Value:   10,
								Usage:   "fetch plans from google calendar",
							},
							&cli.DurationFlag{
								Name:    "within",
								Aliases: []string{"w"},
								Value:   time.Hour * 24 * 14,
								Usage:   "fetch plans within target duration from google calendar",
							},
						},
					},
				},
			},
			{
				Name:  "notify",
				Usage: "simple notify with message",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "locale",
						Aliases: []string{"l"},
						Value:   "en",
					},
					&cli.StringFlag{
						Name:    "message",
						Aliases: []string{"m"},
						Value:   "Hello, world!!",
					},
				},
				Action: notifyFromDevices,
			},
		},
	}

	sort.Sort(cli.FlagsByName(app.Flags))
	sort.Sort(cli.CommandsByName(app.Commands))

	err := app.Run(os.Args)

	if err != nil {
		log.Fatal(err)
	}
}
