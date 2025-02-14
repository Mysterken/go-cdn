package main

import (
  "fmt"
  "log"
  "net/http"
  "os"
  "sync"
  "time"
  "github.com/gin-gonic/gin"
  "github.com/golang-jwt/jwt/v4"
	"go-cdn/routes"
	"path/filepath"
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

	// Start server
	fmt.Println("Starting CDN server on http://localhost:8080")
	log.Fatal(server.ListenAndServe())
}

// Get file from cache or read from disk
func (c *Cache) getFile(path string) ([]byte, error) {
    c.mu.RLock()
    data, found := c.data[path]
    exp, exists := c.expires[path]
    c.mu.RUnlock()

    if found && exists && time.Now().Before(exp) {
        return data, nil
    }

    c.mu.Lock()
    defer c.mu.Unlock()

    // Double-check to avoid redundant file reads
    data, found = c.data[path]
    exp, exists = c.expires[path]
    if found && exists && time.Now().Before(exp) {
        return data, nil
    }

    data, err := os.ReadFile(path)
    if err != nil {
        return nil, err
    }

    c.data[path] = data
    c.expires[path] = time.Now().Add(c.ttl)

    return data, nil
}

var jwtKey = []byte("your_secret_key")

type Claims struct {
    Username string `json:"username"`
    jwt.RegisteredClaims
}

// Generate JWT token
func generateToken(username string) (string, error) {
    expirationTime := time.Now().Add(24 * time.Hour)
    claims := &Claims{
        Username: username,
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(expirationTime),
        },
    }

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString(jwtKey)
}

// Middleware to authenticate JWT tokens
func authenticateJWT() gin.HandlerFunc {
    return func(c *gin.Context) {
        tokenString := c.GetHeader("Authorization")
        if tokenString == "" {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
            c.Abort()
            return
        }

        // Remove "Bearer " prefix if present
        if len(tokenString) > 7 && tokenString[:7] == "Bearer " {
            tokenString = tokenString[7:]
        }

        claims := &Claims{}
        token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
            return jwtKey, nil
        })

        if err != nil || !token.Valid {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
            c.Abort()
            return
        }

        c.Set("username", claims.Username)
        c.Next()
    }
}

// Serve static files with caching
func serveStaticFiles(c *Cache) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        if r.URL.Path == "/health" {
            w.WriteHeader(http.StatusOK)
            _, _ = w.Write([]byte("OK"))
            return
        }

        filePath := "." + r.URL.Path

        // Vérifier si le fichier existe en cache (sans stocker inutilement "data")
        if _, err := c.getFile(filePath); err != nil {
            http.Error(w, "File not found", http.StatusNotFound)
            return
        }

        // Ouvrir le fichier proprement
        file, err := os.Open(filePath)
        if err != nil {
            http.Error(w, "Failed to open file", http.StatusInternalServerError)
            return
        }
        defer file.Close()

        // Détecter le type MIME et servir correctement
        http.ServeContent(w, r, filePath, time.Now(), file)
    }
}

func main() {
    cache := newCache(10 * time.Second)

    // Start the static file server in a goroutine
    go func() {
        mux := http.NewServeMux()
        mux.HandleFunc("/", serveStaticFiles(cache))

        server := &http.Server{
            Addr:         ":8082",
            Handler:      mux,
            ReadTimeout:  5 * time.Second,
            WriteTimeout: 10 * time.Second,
            IdleTimeout:  15 * time.Second,
        }

        fmt.Println("Starting CDN server on http://localhost:8082")
        if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            log.Fatalf("Static file server error: %v", err)
        }
    }()

    // Start the API server
    r := gin.Default()

    // CORS middleware
    r.Use(func(c *gin.Context) {
        c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
        c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
        c.Writer.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization")
        if c.Request.Method == "OPTIONS" {
            c.AbortWithStatus(http.StatusOK)
            return
        }
        c.Next()
    })

    // Login route
    r.POST("/api/login", func(c *gin.Context) {
        var creds struct {
            Username string `json:"username"`
            Password string `json:"password"`
        }
        if err := c.BindJSON(&creds); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
            return
        }

        if creds.Username == "user" && creds.Password == "password" {
            token, err := generateToken(creds.Username)
            if err != nil {
                c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
                return
            }
            c.JSON(http.StatusOK, gin.H{"token": token})
        } else {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
        }
    })

    // Protected route
    r.GET("/api/hello", authenticateJWT(), func(c *gin.Context) {
        username, _ := c.Get("username")
        c.JSON(http.StatusOK, gin.H{"message": "Hello " + username.(string)})
    })

    fmt.Println("Starting API server on http://localhost:8083")
    if err := r.Run(":8083"); err != nil {
        log.Fatalf("API server error: %v", err)
    }
}