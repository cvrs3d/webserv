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
	}
	multiplexer := http.NewServeMux()

	multiplexer.Handle("/app/", apiCfg.middlewareMetrics(http.StripPrefix("/app/", http.FileServer(http.Dir(".")))))

	assetsHandler := http.StripPrefix("/app/assets", http.FileServer(http.Dir("./assets")))
	multiplexer.Handle("/app/assets", apiCfg.middlewareMetrics(assetsHandler))
	multiplexer.Handle("/app/assets/", apiCfg.middlewareMetrics(assetsHandler))
	multiplexer.HandleFunc("GET /api/healthz", healthHandler)
	multiplexer.HandleFunc("GET /admin/metrics", apiCfg.metricsHandler)
	multiplexer.HandleFunc("POST /admin/reset", apiCfg.resetHandler)
	multiplexer.HandleFunc("POST /api/users", apiCfg.usersHandler)
	multiplexer.HandleFunc("POST /api/login", apiCfg.loginHandler)
	multiplexer.HandleFunc("POST /api/chirps", apiCfg.validateHandler)
	multiplexer.HandleFunc("GET /api/chirps", apiCfg.getChirpsHandler)
	multiplexer.HandleFunc("GET /api/chirps/{chirp_id}", apiCfg.getChirpByIDHandler)
	


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
