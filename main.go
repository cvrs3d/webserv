package main

import (
	"log"
	"net/http"
)

func main() {
	multiplexer := http.NewServeMux()

	multiplexer.Handle("/app/", http.StripPrefix("/app/", http.FileServer(http.Dir("."))))

	assetsHandler := http.StripPrefix("/app/assets", http.FileServer(http.Dir("./assets")))
	multiplexer.Handle("/app/assets", assetsHandler)
	multiplexer.Handle("/app/assets/", assetsHandler)
	multiplexer.HandleFunc("/healthz", healthHandler)

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
