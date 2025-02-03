package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"sync"
	"time"
)

type URLStore struct {
	urls  map[string]urlData
	mutex sync.RWMutex
}

type urlData struct {
	longURL   string
	createdAt time.Time
}

type Response struct {
	ShortURL string `json:"short_url"`
	LongURL  string `json:"long_url,omitempty"`
	Error    string `json:"error,omitempty"`
}

type CreateURLRequest struct {
	LongURL string `json:"long_url"`
}

var store = &URLStore{
	urls: make(map[string]urlData),
}

func generateShortURL() string {
	chars := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, 6)
	for i := range result {
		result[i] = chars[rand.Intn(len(chars))]
	}
	return string(result)
}

func cleanup() {
	for {
		time.Sleep(1 * time.Hour)
		now := time.Now()

		store.mutex.Lock()
		for short, data := range store.urls {
			if now.Sub(data.createdAt) > 1*time.Hour {
				delete(store.urls, short)
			}
		}
		store.mutex.Unlock()
	}
}

func createShortURL(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req CreateURLRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.LongURL == "" {
		http.Error(w, "Missing URL", http.StatusBadRequest)
		return
	}

	short := generateShortURL()
	store.mutex.Lock()
	store.urls[short] = urlData{
		longURL:   req.LongURL,
		createdAt: time.Now(),
	}
	store.mutex.Unlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(Response{
		ShortURL: fmt.Sprintf("http://localhost:8080/%s", short),
		LongURL:  req.LongURL,
	})
}

func redirectToLongURL(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	store.mutex.RLock()
	data, ok := store.urls[r.URL.Path[1:]]
	store.mutex.RUnlock()

	if !ok {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	http.Redirect(w, r, data.longURL, http.StatusMovedPermanently)
}

func main() {
	go cleanup()

	http.HandleFunc("/shorten", createShortURL)
	http.HandleFunc("/", redirectToLongURL)

	fmt.Println("Starting server on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Printf("Server failed to start: %v\n", err)
	}
}
