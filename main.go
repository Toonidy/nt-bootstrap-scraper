package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/chromedp/cdproto/dom"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
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
					source, err := downloadBootstrapFile(context.Background())
					if err != nil {
						return fmt.Errorf("unable to downloda bootstrap.js: %w", err)
					}
					fmt.Println("Test", source)
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

func downloadBootstrapFile(ctx context.Context) (string, error) {
	// Setup Chrome
	ctx, cancel := chromedp.NewExecAllocator(ctx,
		chromedp.UserAgent("Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/96.0.4664.110 Safari/537.36"),
		chromedp.NoFirstRun,
		chromedp.NoDefaultBrowserCheck,
		chromedp.NoSandbox,
		chromedp.Headless,
		chromedp.DisableGPU,
	)
	defer cancel()

	ctx, cancel = chromedp.NewContext(
		ctx,
		chromedp.WithLogf(log.Printf),
	)
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 120*time.Second)
	defer cancel()

	// Find bootstrap.js
	var bootstrapSrc string

	err := chromedp.Run(ctx,
		chromedp.Navigate("https://www.nitrotype.com/"),
		chromedp.WaitReady("#root"),
		chromedp.ActionFunc(func(ctx context.Context) error {
			node, err := dom.GetDocument().Do(ctx)
			if err != nil {
				return err
			}
			bootstrapNode, err := dom.QuerySelector(node.NodeID, `script[src$="bootstrap.js"]`).Do(ctx)
			if err != nil {
				return err
			}
			attributes, err := dom.GetAttributes(bootstrapNode).Do(ctx)
			if err != nil {
				return err
			}
			for i, att := range attributes {
				if i%2 != 0 {
					continue
				}
				if att == "src" {
					if i+i < len(attributes) {
						bootstrapSrc = attributes[i+1]
					}
					break
				}
			}
			return nil
		}),
		chromedp.ActionFunc(func(ctx context.Context) error {
			if bootstrapSrc == "" {
				return fmt.Errorf("bootstrap.js not found")
			}
			bootstrapSrc = "https://www.nitrotype.com" + bootstrapSrc
			return nil
		}),
	)
	if err != nil {
		return "", err
	}

	// Setup download
	var requestID network.RequestID
	downloadComplete := make(chan bool)

	chromedp.ListenTarget(ctx, func(v interface{}) {
		switch ev := v.(type) {
		case *network.EventRequestWillBeSent:
			if ev.Request.URL == bootstrapSrc {
				requestID = ev.RequestID
			}
		case *network.EventLoadingFinished:
			if ev.RequestID == requestID {
				close(downloadComplete)
			}
		}
	})

	err = chromedp.Run(ctx, chromedp.Navigate(bootstrapSrc))
	if err != nil {
		return "", err
	}

	<-downloadComplete

	// get the downloaded bytes for the request id
	var downloadBytes []byte
	if err := chromedp.Run(ctx, chromedp.ActionFunc(func(ctx context.Context) error {
		downloadBytes, err = network.GetResponseBody(requestID).Do(ctx)
		return err
	})); err != nil {
		return "", err
	}

	return string(downloadBytes), nil
}
