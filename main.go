package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"sync"
	"time"
)

type URLStore struct {
	sync.RWMutex
	urls map[string]string
}

var store = URLStore{
	urls: make(map[string]string),
}

type ShortenRequest struct {
	URL string `json:"url"`
}

type ShortenResponse struct {
	ShortURL string `json:"short_url"`
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

func generateShortCode() string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	shortCode := make([]byte, 6)
	for i := range shortCode {
		shortCode[i] = charset[rand.Intn(len(charset))]
	}
	return string(shortCode)
}

func shortenURLHandler(w http.ResponseWriter, r *http.Request) {
	var req ShortenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	store.Lock()
	defer store.Unlock()

	shortCode := generateShortCode()
	store.urls[shortCode] = req.URL

	resp := ShortenResponse{
		ShortURL: fmt.Sprintf("http://localhost:8080/%s", shortCode),
	}

	json.NewEncoder(w).Encode(resp)

}

func redirectHandler(w http.ResponseWriter, r *http.Request) {
	shortCode := r.URL.Path[len("/"):]

	store.RLock()
	defer store.RUnlock()

	originalURL, exists := store.urls[shortCode]

	if !exists {
		http.Error(w, "URL not found", http.StatusNotFound)
		return
	}

	http.Redirect(w, r, originalURL, http.StatusFound)

}

func main() {

	http.HandleFunc("/shorten", shortenURLHandler)
	http.HandleFunc("/", redirectHandler)

	log.Println("Starting server on :8080")
	log.Fatal((http.ListenAndServe(":8080", nil)))

}
