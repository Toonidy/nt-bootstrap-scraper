package main

import (
	"flag"
	"log"
	"os"

	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Usage: "runs an api server containing Nitro Type booststrap data.",
		Action: func(c *cli.Context) error {
			return flag.ErrHelp
		},
	}
	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
