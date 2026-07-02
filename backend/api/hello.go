// api/hello.go
package handler

import (
	"fmt"
	"net/http"
)

// Fungsi harus diawali huruf kapital (Handler) agar diexport oleh Vercel
func Handler(w http.ResponseWriter, r *http.Request) {
	// Setup CORS agar bisa ditembak dari frontend Netlify kamu nanti
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Content-Type", "application/json")

	// Tangani preflight request untuk CORS
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Mengembalikan response JSON
	fmt.Fprintf(w, `{"message": "Halo dari Backend Go di Vercel!"}`)
}