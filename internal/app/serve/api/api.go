package api

import (
	"context"
	"encoding/json"
	"encoding/xml"
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

type BootstrapXML struct {
	XMLName    xml.Name      `xml:"bootstrap"`
	TopPlayers TopPlayersXML `xml:"top-players"`
	TopTeams   TopTeamsXML   `xml:"top-teams"`
}

type TopPlayersXML struct {
	Players []RankItemXML `xml:"players"`
}

type TopTeamsXML struct {
	Teams []RankItemXML `xml:"teams"`
}

type RankItemXML struct {
	ID          int `xml:"id,attr"`
	Rank        int `xml:"rank,attr"`
	TopPosition int `xml:"top-position,attr"`
}

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
		r.Get("/bootstrap/xml", func(w http.ResponseWriter, r *http.Request) {
			log := logger.With(zap.String("reqID", middleware.GetReqID(r.Context())))

			source, err := getBootstrapData(cacheManager)
			if err != nil {
				log.Error("grabbing bootstrap data from nitro type failed", zap.Error(err))

				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("Unable to collect NT Bootstrap Data. Please try again later."))
				return
			}

			// Populate Output
			output := &BootstrapXML{}
			if data, ok := source["TOP_PLAYERS"]; ok {
				topPlayers, ok := data.([]nitrotype.RankItem)
				if !ok {
					log.Error("grabbing top players data from nitro type failed", zap.Error(err))

					w.WriteHeader(http.StatusInternalServerError)
					w.Write([]byte("Unable to collect NT Bootstrap Data. Please try again later."))
					return
				}
				for rank, player := range topPlayers {
					rankItem := RankItemXML{
						Rank:        rank + 1,
						ID:          player.ID,
						TopPosition: player.Position,
					}
					output.TopPlayers.Players = append(output.TopPlayers.Players, rankItem)
				}
			}
			if data, ok := source["TOP_TEAMS"]; ok {
				topTeams, ok := data.([]nitrotype.RankItem)
				if !ok {
					log.Error("grabbing top teams data from nitro type failed", zap.Error(err))

					w.WriteHeader(http.StatusInternalServerError)
					w.Write([]byte("Unable to collect NT Bootstrap Data. Please try again later."))
					return
				}
				for rank, team := range topTeams {
					rankItem := RankItemXML{
						Rank:        rank + 1,
						ID:          team.ID,
						TopPosition: team.Position,
					}
					output.TopTeams.Teams = append(output.TopTeams.Teams, rankItem)
				}
			}

			// Output XML
			w.Header().Set("Content-Type", "application/xml")
			enc := xml.NewEncoder(w)
			enc.Indent("  ", "    ")
			if err := enc.Encode(output); err != nil {
				log.Error("exporting bootstrap data from nitro type failed", zap.Error(err))
			}
		})
		r.Get("/bootstrap", func(w http.ResponseWriter, r *http.Request) {
			log := logger.With(zap.String("reqID", middleware.GetReqID(r.Context())))

			source, err := getBootstrapData(cacheManager)
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
	})
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hi?"))
	})

	return r
}

// getBootstrapData fetchces NT Bootstrap Data from the cache or the net
func getBootstrapData(cacheManager *cache.Cache) (nitrotype.NTGLOBALS, error) {
	cacheSource, found := cacheManager.Get("bootstrap_data")
	if found {
		source, ok := cacheSource.(nitrotype.NTGLOBALS)
		if !ok {
			return nil, fmt.Errorf("failed to fetch nitro type bootstrap js from cache")
		}
		return source, nil
	}

	source, err := nitrotype.GetBootstrapData(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to get latest nitro type bootstrap js: %w", err)
	}

	cacheManager.Set("bootstrap_data", source, cache.DefaultExpiration)
	return source, nil
}
