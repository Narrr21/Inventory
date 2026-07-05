package main

import (
	"log"
	"net/http"
	"os"

	"my-backend/internal/handlers"
)

func main() {
	mux := handlers.NewRouter()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Stub API listening on port %s...\n", port)
	log.Fatal(http.ListenAndServe(":"+port, mux))
}
