package main

import (
	"fmt"
	"log"
	"os"

	"gtail/pkg/log_watcher"

	"github.com/urfave/cli"
)

func main() {
	var pattern string
	var isDebug int
	app := &cli.App{
		Name:  "gtail",
		Usage: "golang tail",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "pattern",
				Usage:       "language for the greeting",
				Required:    true,
				Destination: &pattern,
			},
			&cli.IntFlag{
				Name: "verbose",
				Usage: "show log",
				Value: 0,
				Destination: &isDebug,
			},
		},
		Action: func(c *cli.Context) error {
			if len(pattern) == 0 {
				fmt.Printf("miss pattern\n")
				return nil
			}

			var logWatcher = log_watcher.NewLogWatcher(pattern, isDebug)
			logWatcher.Prepare()
			logWatcher.Tail()

			return nil
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
