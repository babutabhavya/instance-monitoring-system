package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

func NewLoggingMiddleware(logger *log.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger.Printf("Request: %s %s %s", r.Method, r.URL.Path, time.Now().Format(time.RFC1123))
		next.ServeHTTP(w, r)
	})
}

type Response struct {
	Message string `json:"message"`
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	response := Response{
		Message: "Healthy!",
	}

	err := json.NewEncoder(w).Encode(response)
	if err != nil {
		http.Error(w, "Unable to encode response", http.StatusInternalServerError)
	}
}

func main() {
	logger := log.New(log.Writer(), "INFO: ", log.LstdFlags)

	router := mux.NewRouter()

	router.HandleFunc("/health", healthHandler).Methods("GET")

	var handler http.Handler = router
	handler = NewLoggingMiddleware(logger, handler)

	log.Println("Starting server on :8080...")
	err := http.ListenAndServe(":8080", handler)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}
