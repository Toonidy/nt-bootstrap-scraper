package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"nt-bootstrap-scraper/pkg/nitrotype"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/httprate"
	"github.com/patrickmn/go-cache"
	"go.uber.org/zap"
)

// NewAPIService sets up the API Service for Raffles
func NewAPIService(logger *zap.Logger, cacheManager *cache.Cache, corsOptions *cors.Options) http.Handler {
	corsMiddleware := cors.Handler(*corsOptions)

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(loggerMiddleware(logger))
	r.Use(corsMiddleware)
	r.Use(httprate.LimitByIP(100, 1*time.Minute))

	r.Route("/api", func(r chi.Router) {
		r.Get("/check", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("OK"))
		})
		r.Get("/bootstrap", func(w http.ResponseWriter, r *http.Request) {
			log := logger.With(zap.String("reqID", middleware.GetReqID(r.Context())))

			source, err := getBootstrapData(logger, cacheManager)
			if err != nil {
				log.Error("grabbing bootstrap data from nitro type failed", zap.Error(err))

				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("Unable to collect NT Bootstrap Data. Please try again later."))
				return
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			err = json.NewEncoder(w).Encode(source)
			if err != nil {
				log.Error("exporting bootstrap data from nitro type failed", zap.Error(err))
			}
		})
		r.Get("/racer/{username}", func(w http.ResponseWriter, r *http.Request) {
			log := logger.With(zap.String("reqID", middleware.GetReqID(r.Context())))

			username := chi.URLParam(r, "username")
			if username == "" {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte("Invalid racer profile request."))
				return
			}

			racer, err := getPlayerData(logger, cacheManager, username)
			if err != nil {
				log.Error("grabbing player data from nitro type failed", zap.Error(err))

				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("Unable to collect NT Player Data. Please try again later."))
				return
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			err = json.NewEncoder(w).Encode(racer)
			if err != nil {
				log.Error("exporting racer data from nitro type failed", zap.Error(err))
			}
		})
	})
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hi?"))
	})

	return r
}

// loggerMiddleware is a middleware that logs the start and end of each request, along
// with some useful data about what was requested, what the response status was,
// and how long it took to return.
func loggerMiddleware(l *zap.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			t1 := time.Now()
			defer func() {
				l.Info("Served",
					zap.String("proto", r.Proto),
					zap.String("path", r.URL.Path),
					zap.Duration("lat", time.Since(t1)),
					zap.Int("status", ww.Status()),
					zap.Int("size", ww.BytesWritten()),
					zap.String("reqId", middleware.GetReqID(r.Context())))
			}()

			next.ServeHTTP(ww, r)
		}
		return http.HandlerFunc(fn)
	}
}

// getPlayerData fetchces NT Bootstrap Data from the cache or the net
func getPlayerData(logger *zap.Logger, cacheManager *cache.Cache, username string) (nitrotype.NTPlayerLegacy, error) {
	cacheName := "player_data_" + username
	cacheSource, found := cacheManager.Get(cacheName)
	if found {
		source, ok := cacheSource.(nitrotype.NTPlayerLegacy)
		if ok {
			return source, nil
		}
		logger.Warn("failed to read player data from cache")
	}
	racer, err := nitrotype.GetPlayerData(context.Background(), username)
	if err != nil {
		return nil, fmt.Errorf("failed to get latest nitro type player data: %w", err)
	}

	cacheManager.Set(cacheName, racer, cache.DefaultExpiration)
	return racer, nil

}

// getBootstrapData fetchces NT Bootstrap Data from the cache or the net
func getBootstrapData(logger *zap.Logger, cacheManager *cache.Cache) (*nitrotype.NTGlobalsLegacy, error) {
	cacheSource, found := cacheManager.Get("bootstrap_data")
	if found {
		source, ok := cacheSource.(*nitrotype.NTGlobalsLegacy)
		if ok {
			return source, nil
		}
		logger.Warn("failed to read bootstrap data from cache")
	}

	source, err := nitrotype.GetBootstrapData(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to get latest nitro type bootstrap js: %w", err)
	}

	cacheManager.Set("bootstrap_data", source, cache.DefaultExpiration)
	return source, nil
}
