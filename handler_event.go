package main

import (
	"encoding/json"
	"net/http"

	"github.com/Tadateki/Chirpy/internal/auth"
	"github.com/Tadateki/Chirpy/internal/database"
	"github.com/google/uuid"
)

func (cfg *apiConfig) eventHandler(w http.ResponseWriter, r *http.Request) {

	//authorization
	apikey, err := auth.GetAPIKey(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "No authorization in hader")
	}

	if apikey != cfg.polka_key {
		respondWithError(w, http.StatusUnauthorized, "API KEY is wrong")
	}

	type UserRequest struct {
		Event string `json:"event"`
		Data  struct {
			UserID string `json:"user_id"`
		} `json:"data"`
	}

	// JSON Purse
	var req UserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	userID, err := uuid.Parse(req.Data.UserID)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid User ID")
		return
	}

	// User Check
	user, err := cfg.dbQueries.GetUserFromUserID(r.Context(), userID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "User ID Not Found")
		return
	}

	// Event Check
	switch req.Event {
	case EventUserUpgraded:

		arg := database.SetChirpyRedParams{
			IsChirpyRed: true,
			ID:          user.ID,
		}

		err := cfg.dbQueries.SetChirpyRed(r.Context(), arg)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "DB Error")
			return
		}
		w.WriteHeader(http.StatusNoContent)

	default:
		w.WriteHeader(http.StatusNoContent)
	}

}
