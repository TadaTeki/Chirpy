package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/Tadateki/Chirpy/internal/auth"
)

func (cfg *apiConfig) refreshHandler(w http.ResponseWriter, r *http.Request) {

	r_token, err := auth.GetBearerToken(r.Header)

	// Authorization Check
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, fmt.Sprintf(`no authorization in header %s"`, err.Error()))
		return
	}

	// DB Lookup for Refresh Token
	refresh_token, err := cfg.dbQueries.GetRefreshTokenFromToken(r.Context(), r_token)
	if err != nil || time.Now().After(refresh_token.ExpiresAt) || refresh_token.RevokedAt.Valid {
		respondWithError(w, http.StatusUnauthorized, "Invalid refresh token")
		return
	}

	user, err := cfg.dbQueries.GetUserFromUserID(r.Context(), refresh_token.UserID)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "no user in database")
		return
	}

	// JWT Token Generation
	token, err := auth.MakeJWT(user.ID, cfg.tokenSecret, time.Duration(cfg.expires_in_seconds)*time.Second)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Access Token Generation Error")
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{
		"token": token,
	})
}
