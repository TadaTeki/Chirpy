package main

import (
	"database/sql"
	"errors"
	"net/http"
	"sort"

	"github.com/Tadateki/Chirpy/internal/database"
	"github.com/google/uuid"
)

const (
	ORDER_ASC = "asc"
	ORDER_DSC = "desc"
)

func (cfg *apiConfig) getchirpsHandler(w http.ResponseWriter, r *http.Request) {
	s := r.URL.Query().Get("author_id")
	var err error
	var user database.User
	var userID uuid.UUID

	// Author ID LOOKUP
	if s != "" {
		userID, err = uuid.Parse(s)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, "invalid User ID")
			return
		}
		// User Check
		user, err = cfg.dbQueries.GetUserFromUserID(r.Context(), userID)
		if err != nil {
			respondWithJSON(w, http.StatusNoContent, "")
			//respondWithError(w, http.StatusNoContent, "No Chirps for the author")
			return
		}

	}

	// Chirps Lookup
	var chirps []database.Chirp

	if s == "" {
		// DBからChirpsを取得
		chirps, err = cfg.dbQueries.GetChirps(r.Context())
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "ERR_DB")
			return
		}
	} else {
		chirps, err = cfg.dbQueries.GetChirpsByAuthor(r.Context(), user.ID)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "ERR_DB")
			return
		}
	}

	// Order by created_at ASC is done in SQL

	order := r.URL.Query().Get("sort")
	if order != ORDER_ASC && order != ORDER_DSC {
		respondWithError(w, http.StatusBadRequest, "Invalid sort parameter; must be 'ASC' or 'DESC'")
	} else if order == ORDER_DSC {
		sort.Slice(chirps, func(i, j int) bool { return chirps[i].CreatedAt.After(chirps[j].CreatedAt) })
	}

	// Chirpsの情報を返す
	var response []map[string]string
	for _, chirp := range chirps {
		response = append(response, map[string]string{
			"id":         chirp.ID.String(),
			"created_at": chirp.CreatedAt.String(),
			"updated_at": chirp.UpdatedAt.String(),
			"body":       chirp.Body,
			"user_id":    chirp.UserID.String(),
		})
	}
	respondWithJSON(w, http.StatusOK, response)

}

func (cfg *apiConfig) getChirpByIDHandler(w http.ResponseWriter, r *http.Request) {
	// URLパスからchirpIDを取得
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
	respondWithJSON(w, http.StatusOK, map[string]string{
		"id":         chirp.ID.String(),
		"created_at": chirp.CreatedAt.String(),
		"updated_at": chirp.UpdatedAt.String(),
		"body":       chirp.Body,
		"user_id":    chirp.UserID.String(),
	})

}
