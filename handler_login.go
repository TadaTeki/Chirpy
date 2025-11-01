package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/Tadateki/Chirpy/internal/auth"
)

func (cfg *apiConfig) loginUserHandler(w http.ResponseWriter, r *http.Request) {
	type LoginUserRequest struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	var req LoginUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid JSON"}`, http.StatusBadRequest)
		return
	}

	user, err := cfg.dbQueries.GetUserFromEmail(r.Context(), req.Email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			respondWithError(w, http.StatusUnauthorized, "Incorrect email or password")
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Database Error")
		return
	}

	chk, err := auth.CheckPasswordHash(req.Password, user.HashedPassword)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Password Check Error")
		return
	}
	if !chk {
		respondWithError(w, http.StatusUnauthorized, "Incorrect email or password")
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{
		"id":         user.ID.String(),
		"created_at": user.CreatedAt.String(),
		"updated_at": user.UpdatedAt.String(),
		"email":      user.Email,
	})

}
