package main

import (
	"sync/atomic"
	"time"

	"github.com/Tadateki/Chirpy/internal/database"
	"github.com/google/uuid"
)

type chirpRequest struct {
	Body string `json:"body"`
}

type createUserRequest struct {
	Email string `json:"email"`
}

type User struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
}

var ngwords = []string{
	"kerfuffle",
	"sharbert",
	"fornax",
}

var maxChirpLength = 140

type apiConfig struct {
	fileserverHits atomic.Int32
	dbQueries      *database.Queries
	platform       string
}
