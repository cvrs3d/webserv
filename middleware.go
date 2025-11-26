package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/cvrs3d/webserv/internal/auth"
	"github.com/cvrs3d/webserv/internal/database"
	"github.com/google/uuid"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	db *database.Queries
	platform string
	secret string
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
		Password string `json:"password"`
	}
	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		log.Printf("Error decoding parameters: %s", err)
		respondWithError(w, 500, "Something went wrong")
		return
	}
	password_hash, _ := auth.HashPassword(params.Password)
	userDTO, err := cfg.db.CreateUser(r.Context(), database.CreateUserParams{
		Email: params.Email,
		HashedPassword: password_hash,
	})
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
	
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		log.Printf("Error getting bearer roken: %s", err)
		respondWithError(w, 401, "Missing header")
		return
	}

	user_id, err := auth.ValidateJWT(token, cfg.secret)
	if err != nil {
		log.Printf("Token not valid: %s", err)
		respondWithError(w, 401, "Auth token is not valid")
		return	
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	if err := decoder.Decode(&params); err != nil {
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
		UserID: user_id,
	})

	if err != nil {
		log.Printf("Error creating Chirp DTO: %s, user_id used: %s", err, user_id.String())
		respondWithError(w, 500, "Something went wrong")
		return
	}

	response := MapChirpDTOToChirp(chirpDTO)

	respondWithJSON(w, 201, response)
}

func (cfg *apiConfig) getChirpsHandler(w http.ResponseWriter, r *http.Request) {
	chirpsDTOS, err := cfg.db.GetChirps(r.Context()) 

	if err != nil {
		log.Printf("Error retrieving chirps: %s", err)
		respondWithError(w, 500, "Something went wrong")
		return
	}

	responseChirps := make([]Chirp, len(chirpsDTOS))

	for i, c := range chirpsDTOS {
		responseChirps[i] = MapChirpDTOToChirp(c)
	}

	respondWithJSON(w, 200, responseChirps)
}

func (cfg *apiConfig) getChirpByIDHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("chirp_id")

	if len(id) == 0 {
		log.Printf("Error getting chirp_id")
		respondWithError(w, 404, "Not found")
		return
	}

	uid, err := uuid.Parse(id)
	if err != nil {
		log.Printf("Error parsing uuid: %s", err)
		respondWithError(w, 500, "Something went wrong")
		return
	}

	chirpDTO, err := cfg.db.GetChirpByID(r.Context(), uid)
	if err != nil {
		log.Printf("Error retrieving chirp: %s", err)
		respondWithError(w, 404, "Nor found")
		return
	}

	response := MapChirpDTOToChirp(chirpDTO)

	respondWithJSON(w, 200, response)
}

func (cfg *apiConfig) loginHandler(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Email string `json:"email"`
		Password string `json:"password"`
		EIS int `json:"expires_in_seconds,omitempty"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		log.Printf("Error decoding parameters: %s", err)
		respondWithError(w, 500, "Something went wrong")
		return
	}
	userDTO, err := cfg.db.GetUserByEmail(r.Context(), params.Email)
	if err != nil {
		log.Printf("Error connecting to db: %s", err)
		respondWithError(w, 500, "Something went wrong")
		return
	}
	if flag, _ := auth.CheckPasswordHash(params.Password, userDTO.HashedPassword); !flag {
		respondWithError(w, 401, "Incorrect email or password")
		return
	}

	if params.EIS < 1 || params.EIS > 60 {
		params.EIS = 60
	}

	jwt, err := auth.MakeJWT(userDTO.ID, cfg.secret, time.Second * time.Duration(params.EIS))

	if err != nil {
		log.Printf("Error constructing the JWT: %s", err)
		respondWithError(w, 500, "Something went wrong")
		return
	}

	refreshToken, _ := auth.MakeRefreshToken()
	parsedToken, _ := uuid.Parse(refreshToken)
	rt, err := cfg.db.CreateRefreshToken(r.Context(), database.CreateRefreshTokenParams{
		Token: parsedToken,
		UserID: userDTO.ID,
		ExpiresAt: time.Now().Add(time.Hour * 24 * time.Duration(60)),
		RevokedAt: sql.NullTime{},
	})
	if err != nil {
		log.Printf("Error constructing the Refresh token: %s", err)
		respondWithError(w, 500, "Something went wrong")
		return
	}

	user := MapUserDTOToUser(userDTO)
	user.JWTToken = jwt
	user.RefreshToken = rt.Token.String()

	respondWithJSON(w, 200, user)
}


func (cfg *apiConfig) refreshHandler(w http.ResponseWriter, r *http.Request) {
	type response struct {
		Token string `json:"token"`
	}
	token, err := auth.GetBearerToken(r.Header)

	if err != nil {
		log.Printf("Error fetching refresh token from a header: %s", err)
		respondWithError(w, 401, "Refresh token is not present")
		return
	}

	parsedToken, _ := uuid.Parse(token)

	tokenDTO, err := cfg.db.GetRefreshTokenByToken(r.Context(), parsedToken)

	if err == sql.ErrNoRows {
		log.Printf("Error refresh token has expired or doesn't exists: %s", err)
		respondWithError(w, 401, "Refresh token has expired or doesn't exists")
		return
	}

	if err != nil {
		log.Printf("Error executing query: %s", err)
		respondWithError(w, 500, "Something went wrong")
		return
	}

	jwt, err := auth.MakeJWT(tokenDTO.UserID, cfg.secret, time.Duration(1) * time.Hour)

	if err != nil {
		log.Printf("Error : %s", err)
		respondWithError(w, 500, "Something went wrong")
		return
	}

	respondWithJSON(w, 200, response {
		Token: jwt,
	})
}

func (cfg *apiConfig) revokeHandler(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header)

	if err != nil {
		log.Printf("Error fetching refresh token from a header: %s", err)
		respondWithError(w, 401, "Refresh token is not present")
		return
	}

	parsedToken, _ := uuid.Parse(token)

	if err := cfg.db.UpdateRefreshToken(r.Context(), parsedToken); err != nil {
		log.Printf("Error fetching refresh token from a database: %s", err)
		respondWithError(w, 401, "Refresh token is right")
		return	
	}

	respondWithJSON(w, 204, struct{}{})
}