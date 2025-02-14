package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Cache structure
type Cache struct {
	mu      sync.RWMutex
	data    map[string][]byte
	expires map[string]time.Time
	ttl     time.Duration
}

func newCache(ttl time.Duration) *Cache {
	return &Cache{
		data:    make(map[string][]byte),
		expires: make(map[string]time.Time),
		ttl:     ttl,
	}
}

// Retrieve a file from the cache or fetch it from the disk
func (c *Cache) getFile(path string) ([]byte, error) {
	c.mu.RLock()
	// Check if file is in cache and not expired
	if data, found := c.data[path]; found {
		if time.Now().Before(c.expires[path]) {
			c.mu.RUnlock()
			return data, nil
		}
	}

	c.mu.RUnlock()

	// Cache miss or expired, read the file from disk
	c.mu.Lock()
	defer c.mu.Unlock()

	// Check again after acquiring the lock
	if data, found := c.data[path]; found {
		if time.Now().Before(c.expires[path]) {
			return data, nil
		}
	}

	// Read from the disk
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	// Store in cache with TTL
	c.data[path] = data
	c.expires[path] = time.Now().Add(c.ttl)

	return data, nil
}

// Handle HTTP requests and serve static files
func serveStaticFiles(c *Cache) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Handle health check separately
		if r.URL.Path == "/health" {
			w.WriteHeader(http.StatusOK)
			if _, err := w.Write([]byte("OK")); err != nil {
				http.Error(w, "Failed to write response", http.StatusInternalServerError)
				return
			}
			return
		}

		// Determine the file path
		filePath := "." + r.URL.Path // assuming files are served from the current directory

		// Try to get file from cache
		data, err := c.getFile(filePath)
		if err != nil {
			http.Error(w, "File not found", http.StatusNotFound)
			return
		}

		// Serve the file content
		w.Header().Set("Content-Type", "application/octet-stream")
		if _, err := w.Write(data); err != nil {
			http.Error(w, "Failed to write response", http.StatusInternalServerError)
			return
		}
	}
}

func main() {
	// Create a cache with a TTL of 10 seconds
	cache := newCache(10 * time.Second)

	// Serve static files with caching
	http.HandleFunc("/", serveStaticFiles(cache))

	// Enregistrer la route pour télécharger cat.jpg
	http.HandleFunc("/upload", DownloadCat)

	// Enregistrer la route dynamique pour télécharger des images
	http.HandleFunc("/download/", DownloadImage)
	// Create a custom server with timeouts

	server := &http.Server{
		Addr:         ":8082",
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  15 * time.Second,
	}

	// Create a context with a timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Set credentials
	credential := options.Credential{
		Username: "root",
		Password: "example",
	}
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017/").SetAuth(credential)

	// Connect to MongoDb
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(ctx)

	// Ping the database to verify the connection
	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatal(err)
	}

	// Check if "filesCache" exists in the list
	var exists = false
	databases, err := client.ListDatabaseNames(ctx, bson.M{})
	if err != nil {
		log.Fatal(err)
	}

	for _, name := range databases {
		if name == "filesCache" {
			exists = true
		}
	}

	if !exists {
		db := client.Database("filesCache")

		files := db.Collection("file")
		file := bson.D{{Key: "pathName", Value: "empty.png"}}

		_, err := files.InsertOne(ctx, file)
		if err != nil {
			log.Fatal(err)
		}
	}

	// Start the server
	fmt.Println("Starting CDN server on http://localhost:8082")
	log.Fatal(server.ListenAndServe())
}
