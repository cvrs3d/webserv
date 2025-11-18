package main

import (
	"log"
	"net/http"
	"sync/atomic"
)


func main() {
	apiCfg := apiConfig{
		fileserverHits: atomic.Int32{},
	}
	multiplexer := http.NewServeMux()

	multiplexer.Handle("/app/", apiCfg.middlewareMetrics(http.StripPrefix("/app/", http.FileServer(http.Dir(".")))))

	assetsHandler := http.StripPrefix("/app/assets", http.FileServer(http.Dir("./assets")))
	multiplexer.Handle("/app/assets", apiCfg.middlewareMetrics(assetsHandler))
	multiplexer.Handle("/app/assets/", apiCfg.middlewareMetrics(assetsHandler))
	multiplexer.HandleFunc("GET /api/healthz", healthHandler)
	multiplexer.HandleFunc("GET /admin/metrics", apiCfg.metricsHandler)
	multiplexer.HandleFunc("POST /admin/reset", apiCfg.resetHandler)

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
