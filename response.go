package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
)

type errorResponse struct {
	Error string `json:"error"`
}

func writeErrorResponse(w http.ResponseWriter, err error, statusCode int) {
	log.Printf("Error decoding parameters: %s", err)

	response := errorResponse{
		Error: err.Error(),
	}
	data, err := json.Marshal(response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	w.Write(data)
}

func writeSuccessResponse(w http.ResponseWriter, jsonMap interface{}, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if jsonMap == nil {
		return
	}

	data, err := json.Marshal(jsonMap)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	w.Write(data)
}

func filterMessage(message string) string {
	words := strings.Split(message, " ")
	cleanWords := make([]string, len(words))
	for i, word := range words {
		loweredWord := strings.ToLower(word)
		if loweredWord == "kerfuffle" || loweredWord == "sharbert" || loweredWord == "fornax" {
			cleanWords[i] = "****"
			continue
		}

		cleanWords[i] = word
	}

	return strings.Join(cleanWords, " ")
}
