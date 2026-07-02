package main

import (
    "net/http"
    "log"
    "my-backend/api" 
)

func main() {
    http.HandleFunc("/api/hello", handler.Handler)
    log.Println("Server lokal berjalan di port 8080...")
    http.ListenAndServe(":8080", nil)
}