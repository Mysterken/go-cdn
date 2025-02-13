package main

import (
	"fmt"
	"go-cdn/routes"
	"log"
	"net/http"
	"time"
)

// Handle HTTP requests and serve static files
func serveStaticFiles(c *Cache) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Health check
		if r.URL.Path == "/health" {
			w.WriteHeader(http.StatusOK)
			if _, err := w.Write([]byte("OK")); err != nil {
				http.Error(w, "Failed to write response", http.StatusInternalServerError)
				return
			}
			return
		}

		// Determine file path
		filePath := "." + r.URL.Path

		// Get file from cache or disk
		data, err := c.getFile(filePath)
		if err != nil {
			http.Error(w, "File not found", http.StatusNotFound)
			return
		}

		// Serve file
		w.Header().Set("Content-Type", "application/octet-stream")
		if _, err := w.Write(data); err != nil {
			http.Error(w, "Failed to write response", http.StatusInternalServerError)
			return
		}
	}
}

func main() {
	// Create an LRU cache with capacity of 100 items and TTL of 30 seconds
	lruCache := newCache(100, 30*time.Second)

	// Serve static files with caching
	http.HandleFunc("/", serveStaticFiles(lruCache))

	// Handler racine personnalisé : si l'URL est "/" on sert index.html,
	// sinon on utilise serveStaticFiles pour les autres fichiers.
	http.HandleFunc("/index.html", serveStaticFiles(cache))

	// Upload a file
	http.HandleFunc("/upload", routes.DownloadCat)

	// Download a file
	http.HandleFunc("/download/", routes.DownloadImage)

	// List files
	http.HandleFunc("/files", fileManager)

	// Create a custom server with timeouts
	server := &http.Server{
		Addr:         ":8080",
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  15 * time.Second,
	}

	// Start server
	fmt.Println("Starting CDN server on http://localhost:8080")
	log.Fatal(server.ListenAndServe())
}
