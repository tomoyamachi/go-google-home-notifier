package main

import (
	"log"
	"os"
	"sort"

	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Commands: []*cli.Command{
			{
				Name:    "daemon",
				Aliases: []string{"d"},
				Usage:   "start notify daemon",
				Action:  startDaemon,
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
						Action:  fetchPlan,
					},
				},
			},
			{
				Name:  "notify",
				Usage: "simple notify",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "message",
						Aliases: []string{"m"},
					},
				},
				Action: notify,
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
