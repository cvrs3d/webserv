package main

import (
	"log"
	"net/http"
)

func main() {
	multiplexer := http.NewServeMux()

	multiplexer.Handle("/", http.FileServer(http.Dir(".")))

	if multiplexer == nil {
		log.Fatal("Allocation error!!!")
	}

	server := http.Server {
		Addr: ":8080",
		Handler: multiplexer,
	}

	defer server.Close()

	if err := server.ListenAndServe(); err != nil {
		log.Fatal("Error: ", err)
	}
}