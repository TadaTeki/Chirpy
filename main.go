package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"sync/atomic"

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

	// 環境変数からPLATFORMを取得
	platform := os.Getenv("PLATFORM")

	// APIサーバー起動
	cfg := &apiConfig{
		fileserverHits: atomic.Int32{},
		dbQueries:      dbQueries,
		platform:       platform,
	}
	servemux := http.NewServeMux()
	servemux.HandleFunc("GET /api/healthz", healthHandler)
	servemux.HandleFunc("GET /admin/metrics", cfg.countHandler)
	servemux.HandleFunc("POST /admin/reset", cfg.resetHandler)
	servemux.HandleFunc("POST /api/chirps", chirpsHandler)
	servemux.HandleFunc("POST /api/users", cfg.createUserHandler)

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

func (cfg *apiConfig) countHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	html := fmt.Sprintf(`
	<html>
		<body>
	      <h1>Welcome, Chirpy Admin</h1>
		      <p>Chirpy has been visited %d times!</p>
		</body>
	</html>Count handler called`, cfg.fileserverHits.Load())
	w.Write([]byte(html))
}

func (cfg *apiConfig) resetHandler(w http.ResponseWriter, r *http.Request) {
	if cfg.platform != "dev" {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	err := cfg.dbQueries.DeleteAllUsers(r.Context())
	if err != nil {
		log.Printf("DeleteAllUsers error: %v", err)
		respondWithError(w, http.StatusInternalServerError, "ERR_DB")
		return
	}

	cfg.fileserverHits.Store(0)

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Reset OK"))
}

func chirpsHandler(w http.ResponseWriter, r *http.Request) {
	// JSONをパース
	var req chirpRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid JSON"}`, http.StatusBadRequest)
		return
	}

	// バリデーション（例: 140文字制限）
	if len(req.Body) > maxChirpLength {
		respondWithError(w, http.StatusBadRequest, "ERR_CHIRP_TOO_LONG")
		return
	}

	// NGワードフィルタリング
	cleaned_body := replaceNGWords(req.Body)
	respondWithJSON(w, http.StatusOK, map[string]string{"cleaned_body": cleaned_body})

}

func (cfg *apiConfig) createUserHandler(w http.ResponseWriter, r *http.Request) {
	// JSONをパース
	var req createUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid JSON"}`, http.StatusBadRequest)
		return
	}

	// DBにユーザーを作成
	user, err := cfg.dbQueries.CreateUser(r.Context(), req.Email)
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

// Hanlder用の内部関数
func respondWithError(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	resp := fmt.Sprintf(`{"error": "%s"}`, message)
	w.Write([]byte(resp))
}

func respondWithJSON(w http.ResponseWriter, cod int, payload interface{}) {
	response, err := json.Marshal(payload)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "JSON Marshal Error")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(cod)
	w.Write(response)
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func middlewareLog(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}

func replaceNGWords(body string) string {

	result := body
	for _, ng := range ngwords {
		re := regexp.MustCompile(`(?i)` + regexp.QuoteMeta(ng))
		result = re.ReplaceAllString(result, "****")
	}
	return result
}
