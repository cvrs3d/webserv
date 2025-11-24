package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync/atomic"

	"github.com/cvrs3d/webserv/internal/database"
	"github.com/google/uuid"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	db *database.Queries
	platform string
}

func (cfg *apiConfig) middlewareMetrics(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func (cfg *apiConfig) metricsHandler(w http.ResponseWriter, r *http.Request) {
	value := cfg.fileserverHits.Load()
	w.Header().Set("Content-type", "text/html")
	message := fmt.Sprintf(`
		<html>
		<body>
			<h1>Welcome, Chirpy Admin</h1>
			<p>Chirpy has been visited %d times!</p>
		</body>
		</html>`, value)
	io.WriteString(w, message)
}

func (cfg *apiConfig) resetHandler(w http.ResponseWriter, r *http.Request) {
	if cfg.platform != "dev" {
		w.WriteHeader(http.StatusForbidden)
		log.Printf("Error: attempted to reset tables on non-dev platform!")
		return
	}
	err := cfg.db.DeleteUsers(r.Context()) 
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("Error reseting users table: %s", err)
		return
	}
	cfg.fileserverHits.Store(0)
	w.WriteHeader(http.StatusOK)
}

func (cfg *apiConfig) usersHandler(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Email string `json:"email"`
	}
	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		log.Printf("Error decoding parameters: %s", err)
		respondWithError(w, 500, "Something went wrong")
		return
	}
	userDTO, err := cfg.db.CreateUser(r.Context(), params.Email)
	if err != nil {
		log.Printf("Error making querry: %s", err)
		respondWithError(w, 500, "Something went wrong")
		return
	}
	user := MapUserDTOToUser(userDTO)
	respondWithJSON(w, 201, user)
}

func (cfg *apiConfig) validateHandler(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Body string `json:"body"`
		UserID uuid.UUID `json:"user_id"`
	}
	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		log.Printf("Error decoding parameters: %s", err)
		respondWithError(w, 500, "Something went wrong")
		return
	}

	if len(params.Body) > 140 {
		respondWithError(w, 400, "Chirp is too long")
		return
	}

	chirpDTO, err := cfg.db.CreateChirp(r.Context(), database.CreateChirpParams{
		Body: params.Body,
		UserID: params.UserID,
	})

	if err != nil {
		log.Printf("Error creating Chirp DTO: %s", err)
		respondWithError(w, 500, "Something went wrong")
		return
	}

	response := MapChirpDTOToChirp(chirpDTO)

	respondWithJSON(w, 201, response)
}