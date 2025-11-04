package main

import (
	"database/sql"
	"errors"
	"net/http"

	"github.com/Tadateki/Chirpy/internal/auth"
	"github.com/google/uuid"
)

func (cfg *apiConfig) deleteChirpyHandler(w http.ResponseWriter, r *http.Request) {

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

	// ChirpID -> chirp
	chirpIDStr := r.PathValue("chirpID")

	chirpID, err := uuid.Parse(chirpIDStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid chirp ID")
		return
	}

	chirp, err := cfg.dbQueries.GetChirp(r.Context(), chirpID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			respondWithError(w, http.StatusNotFound, "chirp not found")
		} else {
			respondWithError(w, http.StatusInternalServerError, "ERR_DB")
		}
		return
	}

	//Check if userid is chirp's userid
	if userid != chirp.UserID {
		respondWithError(w, http.StatusForbidden, "Unathorized Access to chirp")
		return
	}

	err = cfg.dbQueries.DeleteChirp(r.Context(), chirp.ID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Fail to Delete chirp")
		return
	}

	w.WriteHeader(http.StatusNoContent)

}
