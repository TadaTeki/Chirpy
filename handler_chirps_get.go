package main

import (
	"database/sql"
	"errors"
	"log"
	"net/http"

	"github.com/google/uuid"
)

func (cfg *apiConfig) getchirpsHandler(w http.ResponseWriter, r *http.Request) {
	// DBからChirpsを取得
	chirps, err := cfg.dbQueries.GetChirps(r.Context())
	if err != nil {
		log.Printf("GetChirps error: %v", err) // ★追加
		respondWithError(w, http.StatusInternalServerError, "ERR_DB")
		return
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
