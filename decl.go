package main

import (
	"sync/atomic"
	"time"

	"database/sql"

	"github.com/Tadateki/Chirpy/internal/database"
	"github.com/google/uuid"
	//"github.com/vertica/vertica-sql-go/logger"
)

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
	db             *sql.DB
	// logger         *log.Logger
}
