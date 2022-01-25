package cron

import (
	"context"
	"nt-bootstrap-scraper/pkg/nitrotype"

	"github.com/go-logr/zapr"
	"github.com/patrickmn/go-cache"
	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
)

// NewCronService creates a new cron service ready to be activated
func NewCronService(log *zap.Logger, cacheManager *cache.Cache) *cron.Cron {
	logger := zapr.NewLogger(log)
	scrapeBootstrapFN := scrapeBootstrap(log, cacheManager)
	c := cron.New(
		cron.WithChain(cron.DelayIfStillRunning(logger)),
	)
	c.AddFunc("1,11,21,31,41,51 * * * *", scrapeBootstrapFN)

	scrapeBootstrapFN()

	return c
}

// scrapeBootstrap is the scheduled task function that collect Nitro Type Bootstrap file.
func scrapeBootstrap(log *zap.Logger, cacheManager *cache.Cache) func() {
	log = log.With(
		zap.String("job", "scrapeBootstrap"),
	)

	return func() {
		source, err := nitrotype.GetBootstrapData(context.Background())
		if err != nil {
			log.Warn("failed to get latest bootstrap file", zap.Error(err))
		}
		cacheManager.Set("bootstrap_data", source, cache.DefaultExpiration)
		log.Info("bootstrap file updated", zap.Error(err))
	}
}
