package main

import (
	"log"
	"net/http"
	"os"

	"my-backend/api"
)

func main() {
	http.HandleFunc("/api/hello", handler.Handler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // fallback untuk development lokal
	}

	log.Printf("Server berjalan di port %s...\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}