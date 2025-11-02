package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync/atomic"

	"github.com/Tadateki/Chirpy/internal/auth"
	"github.com/Tadateki/Chirpy/internal/database"

	"github.com/joho/godotenv"

	_ "github.com/lib/pq"
)

func main() {

	// .env読み込み
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// DB接続
	dbURL := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal(err)
	}
	dbQueries := database.New(db)
	expires_in_seconds, _ := strconv.Atoi(os.Getenv("EXPIRES_IN_SECONDS"))
	refresh_expires_in_days, _ := strconv.Atoi(os.Getenv("REFRESH_EXPIRES_IN_DAYS"))

	// APIサーバー起動
	cfg := &apiConfig{
		fileserverHits:          atomic.Int32{},
		dbQueries:               dbQueries,
		platform:                os.Getenv("PLATFORM"),
		db:                      db,
		tokenSecret:             os.Getenv("SECRETSTRING"),
		expires_in_seconds:      expires_in_seconds,
		refresh_expires_in_days: refresh_expires_in_days,
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

	servemux.Handle("/app/", cfg.middlewareMetricsInc(http.FileServer(http.Dir("."))))

	server := http.Server{
		Addr:    ":8080",
		Handler: servemux,
	}

	server.ListenAndServe()
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func (cfg *apiConfig) createUserHandler(w http.ResponseWriter, r *http.Request) {
	type createUserRequest struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	// JSONをパース
	var req createUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid JSON"}`, http.StatusBadRequest)
		return
	}

	// PasswordをHash化
	HashedPassword, err := auth.HashPassword(req.Password)
	if err != nil {
		http.Error(w, `{"error":"Hash Password Fail Error"}`, http.StatusBadRequest)
	}

	arg := database.CreateUserParams{
		Email:          req.Email,
		HashedPassword: HashedPassword,
	}

	// DBにユーザーを作成
	user, err := cfg.dbQueries.CreateUser(r.Context(), arg)
	if err != nil {
		log.Printf("CreateUser error: %v", err) // ★追加
		respondWithError(w, http.StatusInternalServerError, "ERR_DB")
		return
	}
	// 作成したユーザーのIDを返す
	respondWithJSON(w, http.StatusCreated, map[string]string{
		"id":         user.ID.String(),
		"created_at": user.CreatedAt.String(),
		"updated_at": user.UpdatedAt.String(),
		"email":      user.Email,
	})

}

func middlewareLog(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}
