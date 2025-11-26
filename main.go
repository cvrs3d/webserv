package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"sync/atomic"

	"github.com/cvrs3d/webserv/internal/database"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)


func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, relying on system env")
	}
	dbURL := os.Getenv("DB_URL")
	platform := os.Getenv("PLATFORM")
	secret := os.Getenv("PRIVATE_KEY")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal(err)
	}
	dbQueries := database.New(db)
	log.Println("Connection established: ", dbQueries)
	apiCfg := apiConfig{
		fileserverHits: atomic.Int32{},
		db: dbQueries,
		platform: platform,
		secret: secret,
	}
	multiplexer := http.NewServeMux()

	multiplexer.Handle("/app/", apiCfg.middlewareMetrics(http.StripPrefix("/app/", http.FileServer(http.Dir(".")))))

	assetsHandler := http.StripPrefix("/app/assets", http.FileServer(http.Dir("./assets")))
	multiplexer.Handle("/app/assets", apiCfg.middlewareMetrics(assetsHandler))
	multiplexer.Handle("/app/assets/", apiCfg.middlewareMetrics(assetsHandler))

	multiplexer.HandleFunc("GET /api/healthz", healthHandler)
	multiplexer.HandleFunc("GET /api/chirps/{chirp_id}", apiCfg.getChirpByIDHandler)
	multiplexer.HandleFunc("GET /admin/metrics", apiCfg.metricsHandler)
	multiplexer.HandleFunc("GET /api/chirps", apiCfg.getChirpsHandler)

	multiplexer.HandleFunc("POST /admin/reset", apiCfg.resetHandler)
	multiplexer.HandleFunc("POST /api/users", apiCfg.usersHandler)
	multiplexer.HandleFunc("POST /api/login", apiCfg.loginHandler)
	multiplexer.HandleFunc("POST /api/refresh", apiCfg.refreshHandler)
	multiplexer.HandleFunc("POST /api/revoke", apiCfg.revokeHandler)
	multiplexer.HandleFunc("POST /api/chirps", apiCfg.validateHandler)

	multiplexer.HandleFunc("PUT /api/users", apiCfg.updateUserHandler)
	
	multiplexer.HandleFunc("DELETE /api/chirps/{chirpID}", apiCfg.deleteChirpByIDHandler)
	


	if multiplexer == nil {
		log.Fatal("Allocation error!!!")
	}

	server := http.Server{
		Addr:    ":8080",
		Handler: multiplexer,
	}

	defer server.Close()

	if err := server.ListenAndServe(); err != nil {
		log.Fatal("Error: ", err)
	}
}
