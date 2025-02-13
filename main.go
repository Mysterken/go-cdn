package main

import (
    "fmt"
    "log"
    "net/http"
    "os"
    "sync"
    "time"
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

    // Create a custom server with timeouts
    server := &http.Server{
        Addr:         ":8080",
        ReadTimeout:  5 * time.Second,
        WriteTimeout: 10 * time.Second,
        IdleTimeout:  15 * time.Second,
    }

    // Start the server
    fmt.Println("Starting CDN server on http://localhost:8080")
    log.Fatal(server.ListenAndServe())
}
