package main

// All code comments should be written in English.

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/Tadateki/Chirpy/internal/auth"
	"github.com/Tadateki/Chirpy/internal/database"
)

func (cfg *apiConfig) loginUserHandler(w http.ResponseWriter, r *http.Request) {
	type LoginUserRequest struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	// JSON purse
	var req LoginUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	// User Lookup
	user, err := cfg.dbQueries.GetUserFromEmail(r.Context(), req.Email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			respondWithError(w, http.StatusUnauthorized, "Incorrect email or password")
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Database Error")
		return
	}

	// Password Check
	chk, err := auth.CheckPasswordHash(req.Password, user.HashedPassword)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Password Check Error")
		return
	}
	if !chk {
		respondWithError(w, http.StatusUnauthorized, "Incorrect email or password")
		return
	}

	// JWT Token Generation
	token, err := auth.MakeJWT(user.ID, cfg.tokenSecret, time.Duration(cfg.expires_in_seconds)*time.Second)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Token Generation Error")
		return
	}

	// Refresh Token Generation
	ref_token_str, err := auth.MakeRefreshToken()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Refresh Token Generation Error")
		return
	}

	arg := database.StoreRefreshTokenParams{
		Token:     ref_token_str,
		UserID:    user.ID,
		ExpiresAt: time.Now().Add(time.Duration(cfg.refresh_expires_in_hours) * time.Hour),
	}

	ref_token, err := cfg.dbQueries.StoreRefreshToken(r.Context(), arg)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Refresh Token Storing Error")
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]any{
		"id":            user.ID.String(),
		"created_at":    user.CreatedAt.String(),
		"updated_at":    user.UpdatedAt.String(),
		"email":         user.Email,
		"token":         token,
		"refresh_token": ref_token.Token,
		"is_chirpy_red": user.IsChirpyRed,
	})

}
