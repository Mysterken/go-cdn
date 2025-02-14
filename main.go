package main

import (
	"fmt"
	"log"
	"net/http"
	"time"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"go-cdn/routes"
)

var jwtKey = []byte("your_secret_key")

type Claims struct {
	Username string `json:"username"`
	jwt.RegisteredClaims
}

// Fonction pour générer un token JWT
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

// Structure pour le cache en mémoire
type Cache struct {
	storage map[string][]byte
}

// Fonction pour créer un nouveau cache
func newCache(size int, ttl time.Duration) *Cache {
	return &Cache{
		storage: make(map[string][]byte, size),
	}
}

// Fonction pour récupérer un fichier depuis le cache
func (c *Cache) getFile(filePath string) ([]byte, error) {
	data, found := c.storage[filePath]
	if !found {
		return nil, fmt.Errorf("fichier non trouvé")
	}
	return data, nil
}

// Fonction pour ajouter un fichier au cache
func (c *Cache) setFile(filePath string, data []byte) {
	c.storage[filePath] = data
}

// Fonction pour gérer les fichiers statiques
func serveStaticFiles(c *Cache) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Vérification de santé
		if r.URL.Path == "/health" {
			w.WriteHeader(http.StatusOK)
			if _, err := w.Write([]byte("OK")); err != nil {
				http.Error(w, "Failed to write response", http.StatusInternalServerError)
				return
			}
			return
		}

		// Déterminer le chemin du fichier
		filePath := "." + r.URL.Path

		// Récupérer le fichier depuis le cache ou le disque
		data, err := c.getFile(filePath)
		if err != nil {
			http.Error(w, "File not found", http.StatusNotFound)
			return
		}

		// Servir le fichier
		w.Header().Set("Content-Type", "application/octet-stream")
		if _, err := w.Write(data); err != nil {
			http.Error(w, "Failed to write response", http.StatusInternalServerError)
			return
		}
	}
}

func main() {
	// Créer un cache
	cache := newCache(100, 10*time.Second)

	// Démarrer le serveur pour les fichiers statiques
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

	// Démarrer le serveur API avec Gin
	r := gin.Default()

	// Route de login
	r.POST("/api/login", func(c *gin.Context) {
		var creds struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}
		if err := c.BindJSON(&creds); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
			return
		}

		// Validation des identifiants
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

	// Route protégée
	r.GET("/api/hello", func(c *gin.Context) {
		username, _ := c.Get("username")
		c.JSON(http.StatusOK, gin.H{"message": "Hello " + username.(string)})
	})

	// Démarrer le serveur API sur http://localhost:8080
	fmt.Println("Starting API server on http://localhost:8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("API server error: %v", err)
	}
}
