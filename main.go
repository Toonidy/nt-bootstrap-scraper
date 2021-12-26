package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"nt-bootstrap-scraper/pkg/nitrotype"
	"os"

	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Usage: "runs an api server containing nitro type booststrap data.",
		Action: func(c *cli.Context) error {
			return flag.ErrHelp
		},
		Commands: []*cli.Command{
			{
				Name:    "collect",
				Aliases: []string{"c"},
				Usage:   "grabs the latest nitro type bootstrap file data.",
				Action: func(c *cli.Context) error {
					source, err := nitrotype.GetBootstrapData(context.Background())
					if err != nil {
						return fmt.Errorf("unable to download bootstrap.js: %w", err)
					}
					output, err := json.Marshal(&source)
					if err != nil {
						return fmt.Errorf("unable to marshal to json: %w", err)
					}
					fmt.Println(string(output))
					return nil
				},
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
