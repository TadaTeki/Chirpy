package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/Tadateki/Chirpy/internal/auth"
)

func (cfg *apiConfig) revokeHandler(w http.ResponseWriter, r *http.Request) {

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

	err = cfg.dbQueries.SetRevokeRefreshToken(r.Context(), refresh_token.Token)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Fail to Revoke refresh token")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
