package main

import (
	"context"
	"fmt"
	"go-cdn/routes"
	"log"
	"net/http"
	"path/filepath"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
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

	// Handler racine personnalisé : si l'URL est "/" ou "/index.html", on sert index.html,
	// sinon on sert les autres fichiers via serveStaticFiles.
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" || r.URL.Path == "/index.html" {
			imageName := filepath.Base(r.URL.Path[len("index.html"):])
			filePath := filepath.Join("", imageName)
			http.ServeFile(w, r, filePath)
			return
		}
		serveStaticFiles(lruCache)(w, r)
	})
	// Enregistrer la route pour télécharger cat.jpg
	http.HandleFunc("/api/upload", routes.DownloadCat)

	// Download a file
	http.HandleFunc("/api/download/", routes.DownloadImage)

	// Create a custom server with timeouts

	// Route pour créer, modifier ou supprimer des fichiers
	http.HandleFunc("/api/files", routes.FileManager)

	server := &http.Server{
		Addr:         ":8080",
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
	fmt.Println("Starting CDN server on http://localhost:8080")
	log.Fatal(server.ListenAndServe())
}
