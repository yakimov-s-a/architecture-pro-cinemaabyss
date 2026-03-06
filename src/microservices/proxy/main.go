package main

import (
	"encoding/json"
	"log"
	"math/rand/v2"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strconv"
)

type Config struct {
	Port                   string
	MonolithUrl            *url.URL
	GradualMigration       bool
	MoviesServiceUrl       *url.URL
	MoviesMigrationPercent int
	EventsServiceUrl       *url.URL
	EventsMigrationPercent int
}

func main() {
	config, err := getConfig()
	if err != nil {
		log.Fatal(err)
	}

	http.Handle("/", &httputil.ReverseProxy{
		Rewrite: func(r *httputil.ProxyRequest) {
			r.SetXForwarded()
			r.SetURL(config.MonolithUrl)
		},
	})

	if config.GradualMigration {
		http.Handle("/api/movies", newWeightedReverseProxy(
			config.MonolithUrl,
			config.MoviesServiceUrl,
			config.MoviesMigrationPercent,
		))
		http.Handle("/api/events", newWeightedReverseProxy(
			config.MonolithUrl,
			config.EventsServiceUrl,
			config.EventsMigrationPercent,
		))
	}

	http.HandleFunc("/health", handleHealth)
	log.Fatal(http.ListenAndServe(":"+config.Port, nil))
}

func getConfig() (*Config, error) {
	port := getEnv("PORT", "8000")

	monolithUrl, err := url.Parse(getEnv("MONOLITH_URL", "http://monolith:8080"))
	if err != nil {
		return nil, err
	}

	gradualMigration, err := strconv.ParseBool(getEnv("GRADUAL_MIGRATION", "true"))
	if err != nil {
		return nil, err
	}

	moviesServiceUrl, err := url.Parse(getEnv("MOVIES_SERVICE_URL", "http://movies-service:8081"))
	if err != nil {
		return nil, err
	}

	moviesMigrationPercent, err := strconv.Atoi(getEnv("MOVIES_MIGRATION_PERCENT", "50"))
	if err != nil {
		return nil, err
	}

	eventsServiceUrl, err := url.Parse(getEnv("EVENTS_SERVICE_URL", "http://events-service:8082"))
	if err != nil {
		return nil, err
	}

	eventsMigrationPercent, err := strconv.Atoi(getEnv("EVENTS_MIGRATION_PERCENT", "50"))
	if err != nil {
		return nil, err
	}

	return &Config{
		Port:                   port,
		MonolithUrl:            monolithUrl,
		GradualMigration:       gradualMigration,
		MoviesServiceUrl:       moviesServiceUrl,
		MoviesMigrationPercent: moviesMigrationPercent,
		EventsServiceUrl:       eventsServiceUrl,
		EventsMigrationPercent: eventsMigrationPercent,
	}, nil
}

func getEnv(key string, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func newWeightedReverseProxy(monolithUrl *url.URL, serviceUrl *url.URL, migrationPercent int) *httputil.ReverseProxy {
	return &httputil.ReverseProxy{
		Rewrite: func(r *httputil.ProxyRequest) {
			r.SetXForwarded()
			if rand.IntN(100) >= migrationPercent {
				r.SetURL(monolithUrl)
			} else {
				r.SetURL(serviceUrl)
			}
		},
	}
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]bool{"status": true}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
