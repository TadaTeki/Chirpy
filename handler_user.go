package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/Tadateki/Chirpy/internal/auth"
	"github.com/Tadateki/Chirpy/internal/database"
)

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
	respondWithJSON(w, http.StatusCreated, map[string]any{
		"id":            user.ID.String(),
		"created_at":    user.CreatedAt.String(),
		"updated_at":    user.UpdatedAt.String(),
		"email":         user.Email,
		"is_chirpy_red": user.IsChirpyRed,
	})

}

func (cfg *apiConfig) updateUserHandler(w http.ResponseWriter, r *http.Request) {

	// Authorization
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "no authorization in header")
		return
	}

	userid, err := auth.ValidateJWT(token, cfg.tokenSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "no authorization in header")
		return
	}

	type createUserRequest struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	// JSONをパース
	var req createUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	// PasswordをHash化
	HashedPassword, err := auth.HashPassword(req.Password)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Hash Password Fail")
	}

	arg := database.UpdateUserParams{
		Email:          req.Email,
		HashedPassword: HashedPassword,
		ID:             userid,
	}

	err = cfg.dbQueries.UpdateUser(r.Context(), arg)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Fail to Update user")
		return
	}

	user, err := cfg.dbQueries.GetUserFromUserID(r.Context(), userid)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Fail to Load updated information")
		return
	}

	// 作成したユーザーのIDを返す
	respondWithJSON(w, http.StatusOK, map[string]any{
		"id":            user.ID.String(),
		"created_at":    user.CreatedAt.String(),
		"updated_at":    user.UpdatedAt.String(),
		"email":         user.Email,
		"is_chirpy_red": user.IsChirpyRed,
	})
}
