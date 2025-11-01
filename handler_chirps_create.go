package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"regexp"
	"time"

	"github.com/Tadateki/Chirpy/internal/database"
	"github.com/google/uuid"
)

func (cfg *apiConfig) chirpsHandler(w http.ResponseWriter, r *http.Request) {
	type createChirpRequest struct {
		Body   string `json:"body"`
		UserID string `json:"user_id"`
	}
	// JSONをパース
	var req createChirpRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid JSON"}`, http.StatusBadRequest)
		return
	}
	if req.Body == "" || req.UserID == "" {
		http.Error(w, `{"error":"missing body or user_id"}`, http.StatusBadRequest)
		return
	}

	userUUID, err := uuid.Parse(req.UserID)
	if err != nil {
		http.Error(w, `{"error":"invalid user_id format"}`, http.StatusBadRequest)
		return
	}

	// バリデーション（例: 140文字制限）
	if len(req.Body) > maxChirpLength {
		respondWithError(w, http.StatusBadRequest, "ERR_CHIRP_TOO_LONG")
		return
	}

	// NGワードフィルタリング
	cleaned_body := replaceNGWords(req.Body)
	//respondWithJSON(w, http.StatusOK, map[string]string{"cleaned_body": cleaned_body})

	// DB に挿入（sqlc で CreateChirp(body, user_id) を生成している前提）
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	chirp, err := cfg.dbQueries.CreateChirp(ctx, database.CreateChirpParams{
		Body:   cleaned_body,
		UserID: userUUID,
	})
	if err != nil {
		log.Printf("CreateChirp error: %v", err) // ★追加
		respondWithError(w, http.StatusInternalServerError, "ERR_DB")
		return
	}

	// 作成した Chirp の情報を返す
	respondWithJSON(w, http.StatusCreated, map[string]string{
		"id":         chirp.ID.String(),
		"created_at": chirp.CreatedAt.String(),
		"updated_at": chirp.UpdatedAt.String(),
		"body":       chirp.Body,
		"user_id":    chirp.UserID.String(),
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
