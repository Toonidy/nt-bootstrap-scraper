package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"nt-bootstrap-scraper/internal/app/serve/api"
	"nt-bootstrap-scraper/pkg/nitrotype"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/go-chi/cors"
	"github.com/patrickmn/go-cache"
	"github.com/urfave/cli/v2"
	"go.uber.org/zap"
)

func main() {
	app := &cli.App{
		Usage: "runs an api server containing nitro type booststrap data.",
		Action: func(c *cli.Context) error {
			return flag.ErrHelp
		},
		Commands: []*cli.Command{
			{
				Name:    "serve",
				Aliases: []string{"s"},
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:    "prod",
						Value:   false,
						Usage:   "whether this is running on production environment",
						EnvVars: []string{"PROD"},
					},
					&cli.StringFlag{
						Name:    "api_addr",
						Value:   ":8080",
						Usage:   "api addr for this server to listen to",
						EnvVars: []string{"API_ADDR"},
					},
					&cli.StringFlag{
						Name:    "cors_allowed_origins",
						Value:   "*",
						Usage:   "allowed origins to access CORS",
						EnvVars: []string{"CORS_ALLOWED_ORIGINS"},
					},
					&cli.StringFlag{
						Name:    "cors_allowed_methods",
						Value:   "GET,POST,PUT,DELETE,OPTIONS",
						Usage:   "allowed http methods for CORS",
						EnvVars: []string{"CORS_ALLOWED_METHODS"},
					},
					&cli.StringFlag{
						Name:    "cors_allowed_headers",
						Value:   "Accept,Authorization,Cache-Control,Content-Type,DNT,If-Modified-Since,Keep-Alive,Origin,User-Agent,X-Requested-With",
						Usage:   "allowed http headers for CORS",
						EnvVars: []string{"CORS_ALLOWED_HEADERS"},
					},
					&cli.BoolFlag{
						Name:    "cors_allow_credentials",
						Value:   true,
						Usage:   "whether to allow credentials for CORS",
						EnvVars: []string{"CORS_ALLOW_CREDENTIALS"},
					},
					&cli.IntFlag{
						Name:    "cors_max_age",
						Value:   1728000,
						Usage:   "TTL to cache CORS",
						EnvVars: []string{"CORS_ALLOW_CREDENTIALS"},
					},
				},
				Usage: "runs a mini api server to serve nitro type boostrap file data.",
				Action: func(c *cli.Context) error {
					cacheManager := cache.New(10*time.Minute, 15*time.Minute)

					corsOptions := &cors.Options{
						AllowedOrigins:   strings.Split(c.String("cors_allowed_origins"), ","),
						AllowedMethods:   strings.Split(c.String("cors_allowed_methods"), ","),
						AllowedHeaders:   strings.Split(c.String("cors_allowed_headers"), ","),
						AllowCredentials: c.Bool("cors_allow_credentials"),
						MaxAge:           c.Int("cors_max_age"),
					}

					apiAddr := c.String("api_addr")

					var logger *zap.Logger

					if c.Bool("prod") {
						logger, _ = zap.NewProduction()
					} else {
						logger, _ = zap.NewDevelopment()
					}
					defer logger.Sync()

					apiService := api.NewAPIService(logger, cacheManager, corsOptions)

					server := &http.Server{
						Addr:    apiAddr,
						Handler: apiService,
					}
					go func() {
						if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
							logger.Fatal("api - service failed to start", zap.Error(err))
						}
					}()
					logger.Info("api - service started")
					logger.Sugar().Infof("api - hosting on %s", apiAddr)

					quit := make(chan os.Signal, 1)
					signal.Notify(quit, os.Interrupt)
					sig := <-quit
					logger.Info("shutting down server...", zap.Any("reason", sig))

					if err := server.Shutdown(context.Background()); err != nil {
						logger.Fatal("api - service failed to shutdown", zap.Error(err))
					}

					return nil
				},
			},
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
