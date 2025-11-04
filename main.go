package main

// All code comments should be written in English.

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync/atomic"

	"github.com/Tadateki/Chirpy/internal/database"

	"github.com/joho/godotenv"

	_ "github.com/lib/pq"
)

const (
	EventUserUpgraded = "user.upgraded"
)

func main() {

	// .env Read
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// DB Connect
	dbURL := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal(err)
	}
	dbQueries := database.New(db)
	expires_in_seconds, _ := strconv.Atoi(os.Getenv("EXPIRES_IN_SECONDS"))
	refresh_expires_in_hours, _ := strconv.Atoi(os.Getenv("REFRESH_EXPIRES_IN_HOURS"))

	// API server setup
	cfg := &apiConfig{
		fileserverHits:           atomic.Int32{},
		dbQueries:                dbQueries,
		platform:                 os.Getenv("PLATFORM"),
		db:                       db,
		tokenSecret:              os.Getenv("SECRETSTRING"),
		expires_in_seconds:       expires_in_seconds,
		refresh_expires_in_hours: refresh_expires_in_hours,
		polka_key:                os.Getenv("POLKA_KEY"),
	}
	servemux := http.NewServeMux()
	servemux.HandleFunc("GET /api/healthz", healthHandler)
	servemux.HandleFunc("GET /admin/metrics", cfg.countHandler)
	servemux.HandleFunc("GET /api/chirps", cfg.getchirpsHandler)
	servemux.HandleFunc("GET /api/chirps/{chirpID}", cfg.getChirpByIDHandler)

	servemux.HandleFunc("POST /admin/reset", cfg.resetHandler)
	servemux.HandleFunc("POST /api/chirps", cfg.chirpsHandler)
	servemux.HandleFunc("POST /api/users", cfg.createUserHandler)
	servemux.HandleFunc("POST /api/login", cfg.loginUserHandler)
	servemux.HandleFunc("POST /api/refresh", cfg.refreshHandler)
	servemux.HandleFunc("POST /api/revoke", cfg.revokeHandler)
	servemux.HandleFunc("POST /api/polka/webhooks", cfg.eventHandler)

	servemux.HandleFunc("DELETE /api/chirps/{chirpID}", cfg.deleteChirpyHandler)

	servemux.HandleFunc("PUT /api/users", cfg.updateUserHandler)

	servemux.Handle("/app/", cfg.middlewareMetricsInc(http.FileServer(http.Dir("."))))

	server := http.Server{
		Addr:    ":8080",
		Handler: servemux,
	}

	server.ListenAndServe()
}

func middlewareLog(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}
