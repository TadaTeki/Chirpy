package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// Hanlder用の内部関数
func respondWithError(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	resp := fmt.Sprintf(`{"error": "%s"}`, message)
	w.Write([]byte(resp))
}

func respondWithJSON(w http.ResponseWriter, cod int, payload interface{}) {
	response, err := json.Marshal(payload)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "JSON Marshal Error")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(cod)
	w.Write(response)
}
