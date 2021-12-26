package api

import (
	"context"
	"encoding/json"
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
	r.Use(corsMiddleware)
	r.Use(httprate.LimitByIP(100, 1*time.Minute))

	r.Route("/api", func(r chi.Router) {
		r.Get("/check", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("OK"))
		})
		r.Get("/bootstrap", func(w http.ResponseWriter, r *http.Request) {
			log := logger.With(zap.String("reqID", middleware.GetReqID(r.Context())))

			source, found := cacheManager.Get("bootstrap_data")
			if found {
				log.Info("returning bootstrap cache")
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				err := json.NewEncoder(w).Encode(source)
				if err != nil {
					log.Error("exporting bootstrap data from nitro type failed", zap.Error(err))
					json.NewEncoder(w).Encode("Unable to export NT Bootstrap Data. Please try again later.")
				}
				return
			}

			source, err := nitrotype.GetBootstrapData(context.Background())
			if err != nil {
				log.Error("grabbing bootstrap data from nitro type failed", zap.Error(err))

				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("Unable to collect NT Bootstrap Data. Please try again later."))
				return
			}

			cacheManager.Set("bootstrap_data", source, cache.DefaultExpiration)

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			err = json.NewEncoder(w).Encode(source)
			if err != nil {
				log.Error("exporting bootstrap data from nitro type failed", zap.Error(err))
				json.NewEncoder(w).Encode("Unable to export NT Bootstrap Data. Please try again later.")
			}
		})
	})
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hi?"))
	})

	return r
}
