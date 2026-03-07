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

func main() {
	port := getEnv("PORT", "8000")

	monolithUrl, err := url.Parse(getEnv("MONOLITH_URL", "http://monolith:8080"))
	if err != nil {
		log.Fatal(err)
	}

	gradualMigration, err := strconv.ParseBool(getEnv("GRADUAL_MIGRATION", "true"))
	if err != nil {
		log.Fatal(err)
	}

	moviesServiceUrl, err := url.Parse(getEnv("MOVIES_SERVICE_URL", "http://movies-service:8081"))
	if err != nil {
		log.Fatal(err)
	}

	moviesMigrationPercent, err := strconv.Atoi(getEnv("MOVIES_MIGRATION_PERCENT", "50"))
	if err != nil {
		log.Fatal(err)
	}

	http.Handle("/", monolithHandler(monolithUrl))
	http.HandleFunc("/health", handleHealth)

	if gradualMigration {
		http.Handle("/api/movies", moviesHandler(monolithUrl, moviesServiceUrl, moviesMigrationPercent))
	}

	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func getEnv(key string, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func monolithHandler(monolithUrl *url.URL) *httputil.ReverseProxy {
	return &httputil.ReverseProxy{
		Rewrite: func(r *httputil.ProxyRequest) {
			r.SetXForwarded()
			r.SetURL(monolithUrl)
		},
	}
}

func moviesHandler(monolithUrl *url.URL, moviesServiceUrl *url.URL, moviesMigrationPercent int) *httputil.ReverseProxy {
	return &httputil.ReverseProxy{
		Rewrite: func(r *httputil.ProxyRequest) {
			r.SetXForwarded()
			if rand.IntN(100) >= moviesMigrationPercent {
				r.SetURL(monolithUrl)
			} else {
				r.SetURL(moviesServiceUrl)
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
