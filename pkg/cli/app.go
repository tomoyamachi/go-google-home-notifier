package cli

import (
	"sort"
	"time"

	"github.com/urfave/cli/v2"
)

var (
	notifyFlags = []cli.Flag{
		&cli.StringFlag{
			Name:  "device-name",
			Usage: "Target Google Home device name. Default notify from all found devices",
		},
		&cli.IntFlag{
			Name:  "device-count",
			Value: 4,
			Usage: "Maximum number of detected Google Home devices",
		},
		&cli.StringFlag{
			Name:    "locale",
			Aliases: []string{"l"},
			Value:   "en",
			Usage:   "Locale code of notifications",
		},
		&cli.StringFlag{
			Name:  "path",
			Value: "",
			Usage: "a Directory path name of credential files (credentials.json, tokens.json)",
		},
	}

	serverFlags = []cli.Flag{
		&cli.IntFlag{
			Name:    "port",
			Aliases: []string{"p"},
			Value:   8000,
		},
	}
)

func App() *cli.App {
	app := &cli.App{
		Commands: []*cli.Command{
			{
				Name:    "daemon",
				Aliases: []string{"d"},
				Usage:   "Start daemon (run server and check calendars regularly)",
				Flags: append(
					append(notifyFlags, serverFlags...),
					&cli.DurationFlag{
						Name:    "notify-duration",
						Aliases: []string{"n"},
						Value:   time.Minute * 30,
						Usage:   "Interval between fetch plans and notify",
					},
					&cli.DurationFlag{
						Name:    "within",
						Aliases: []string{"w"},
						Value:   time.Hour * 2,
						Usage:   "Fetch plans within target duration from Google Calendars",
					},
				),
				Action: startDaemon,
			},
			{
				Name:    "calendar",
				Aliases: []string{"c"},
				Usage:   "About google calendar",
				Subcommands: []*cli.Command{
					{
						Name:    "add-token",
						Aliases: []string{"a"},
						Usage:   "Register a new Google Calendar account",
						Action:  addToken,
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:    "path",
								Aliases: []string{"p"},
								Value:   "",
								Usage:   "a Directory path name of credential files (credentials.json, tokens.json)",
							},
						},
					},
					{
						Name:    "fetch-plan",
						Aliases: []string{"f"},
						Usage:   "Fetch from registered Google Calendars",
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
							&cli.StringFlag{
								Name:    "path",
								Aliases: []string{"p"},
								Value:   "",
								Usage:   "a Directory path name of credential files (credentials.json, tokens.json)",
							},
						},
					},
				},
			},
			{
				Name:  "notify",
				Usage: "Notify a message",
				Flags: append(notifyFlags,
					&cli.StringFlag{
						Name:    "message",
						Aliases: []string{"m"},
						Value:   "Hello, world!!",
					},
				),
				Action: notifyFromDevices,
			},
			{
				Name:   "server",
				Usage:  "Run server",
				Flags:  serverFlags,
				Action: simpleServe,
			},
		},
	}

	sort.Sort(cli.FlagsByName(app.Flags))
	sort.Sort(cli.CommandsByName(app.Commands))
	return app
}
