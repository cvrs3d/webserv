package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"slices"
	"strings"
)



func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8") // normal header
	w.WriteHeader(http.StatusOK)
	io.WriteString(w, "OK")
}

func respondWithError(w http.ResponseWriter, code int, msg string) {
	type responseVal struct {
		Error string `json:"error"`
	}
	response := responseVal {
		Error: msg,
	}
	data, err := json.Marshal(response)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		w.WriteHeader(501)
		return
	}
	w.WriteHeader(code)
	w.Write(data)
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	data, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		w.WriteHeader(500)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(data)
}

func cleanBody(b string) string {
	profane_words := []string{"kerfuffle", "sharbert", "fornax"}
	lowerString := strings.ToLower(b)
	words := strings.Fields(lowerString)
	normalWords := strings.Fields(b)
	out := make([]string, len(words))
	for i, word := range words {
		if slices.Contains(profane_words, word) {
			out[i] = "****"
		} else {
			out[i] = normalWords[i]
		}
	}
	return strings.Join(out, " ")
}