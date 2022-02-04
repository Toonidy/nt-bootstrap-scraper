package nitrotype

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"time"

	"github.com/chromedp/cdproto/dom"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
)

var (
	TopPlayerRegExp          = regexp.MustCompile(`\["TOP_PLAYERS",\{"users":(.*?),"teams":(.*?)\],`)
	TopPlayerMapRegExp       = regexp.MustCompile(`"([0-9]+)":([0-9]+)`)
	UserProfileExtractRegExp = regexp.MustCompile(`(?m)RACER_INFO: (.*),$`)
	ErrPlayerNotFound        = fmt.Errorf("player not found")
)

// GetBootstrapData retrives the NTGLOBALS variable from Nitro Type.
// This function will also manually sort in Top Players and Teams.
func GetBootstrapData(ctx context.Context) (*NTGlobalsLegacy, error) {
	// Setup Chrome
	ctx, cancel := chromedp.NewExecAllocator(ctx,
		chromedp.UserAgent("Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/97.0.4692.99 Safari/537.36"),
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
	var (
		bootstrapSrc string
		// ntGlobals    map[string]interface{}
		ntGlobals NTGlobalsLegacy
	)

	err := chromedp.Run(ctx,
		chromedp.Navigate("https://www.nitrotype.com/"),
		chromedp.WaitReady("#root"),
		chromedp.Evaluate("window.NTGLOBALS", &ntGlobals, chromedp.EvalAsValue),
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

			// Grab Bootstrap Source to manually parse in the Top Players+Teams
			if bootstrapSrc == "" {
				return fmt.Errorf("bootstrap.js not found")
			}
			bootstrapSrc = "https://www.nitrotype.com" + bootstrapSrc

			return nil
		}),
	)
	if err != nil {
		return nil, err
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
		return nil, err
	}

	<-downloadComplete

	// Get the downloaded bytes for the request id
	var downloadBytes []byte
	if err := chromedp.Run(ctx, chromedp.ActionFunc(func(ctx context.Context) error {
		downloadBytes, err = network.GetResponseBody(requestID).Do(ctx)
		return err
	})); err != nil {
		return nil, err
	}

	data := string(downloadBytes)

	// Extract Top Players and Teams in ordered fashion
	topPlayerData := TopPlayerRegExp.FindStringSubmatch(data)
	if len(topPlayerData) != 3 {
		return nil, fmt.Errorf("unable to parse top players")
	}

	var topPlayers []RankItem
	playerData := TopPlayerMapRegExp.FindAllStringSubmatch(topPlayerData[1], -1)
	for i, r := range playerData {
		userID, err := strconv.Atoi(r[1])
		if err != nil {
			return nil, fmt.Errorf("unable to parse top player id (row: %d): %w", i, err)
		}
		position, err := strconv.Atoi(r[2])
		if err != nil {
			return nil, fmt.Errorf("unable to parse top player position (row: %d): %w", i, err)
		}

		topPlayers = append(topPlayers, RankItem{
			ID:       userID,
			Position: position,
		})
	}
	var topTeams []RankItem
	teamData := TopPlayerMapRegExp.FindAllStringSubmatch(topPlayerData[2], -1)
	for i, r := range teamData {
		teamID, err := strconv.Atoi(r[1])
		if err != nil {
			return nil, fmt.Errorf("unable to parse top team id (row: %d): %w", i, err)
		}
		position, err := strconv.Atoi(r[2])
		if err != nil {
			return nil, fmt.Errorf("unable to parse top team position (row: %d): %w", i, err)
		}

		topTeams = append(topTeams, RankItem{
			ID:       teamID,
			Position: position,
		})
	}
	delete(ntGlobals, "TOP_PLAYERS")
	ntGlobals["TOP_PLAYERS"] = topPlayers
	ntGlobals["TOP_TEAMS"] = topTeams
	// ntGlobals.TopPlayers = topPlayers
	// ntGlobals.TopTeams = topTeams

	return &ntGlobals, nil
}

// GetPlayerData fetches the RACER_INFO data from racer profile page.
func GetPlayerData(ctx context.Context, username string) (*NTPlayer, error) {
	// Setup Chrome
	ctx, cancel := chromedp.NewExecAllocator(ctx,
		chromedp.UserAgent("Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/97.0.4692.99 Safari/537.36"),
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

	profileURL := "https://www.nitrotype.com/racer/" + username

	// Setup download
	var requestID network.RequestID
	downloadComplete := make(chan bool)

	chromedp.ListenTarget(ctx, func(v interface{}) {
		switch ev := v.(type) {
		case *network.EventRequestWillBeSent:
			if ev.Request.URL == profileURL {
				requestID = ev.RequestID
			}
		case *network.EventLoadingFinished:
			if ev.RequestID == requestID {
				close(downloadComplete)
			}
		}
	})

	err := chromedp.Run(ctx, chromedp.Navigate("view-source:"+profileURL))
	if err != nil {
		if err.Error() == "encountered an undefined value" {
			return nil, ErrPlayerNotFound
		}
		return nil, err
	}

	<-downloadComplete

	// Get the downloaded bytes for the request id
	var downloadBytes []byte
	if err := chromedp.Run(ctx, chromedp.ActionFunc(func(ctx context.Context) error {
		downloadBytes, err = network.GetResponseBody(requestID).Do(ctx)
		return err
	})); err != nil {
		return nil, err
	}

	matches := UserProfileExtractRegExp.FindSubmatch(downloadBytes)
	if len(matches) != 2 {
		return nil, ErrPlayerNotFound
	}

	var output NTPlayer
	if err := json.Unmarshal(matches[1], &output); err != nil {
		return nil, err
	}

	return &output, nil
}
