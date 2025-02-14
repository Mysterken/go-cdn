// package main

// import (
//     "fmt"
//     "log"
//     "net/http"
//     "os"
//     "sync"
//     "time"

//     "github.com/gin-gonic/gin"
//     "github.com/dgrijalva/jwt-go"
// )

// // Cache structure
// type Cache struct {
//     mu      sync.RWMutex
//     data    map[string][]byte
//     expires map[string]time.Time
//     ttl     time.Duration
// }

// func newCache(ttl time.Duration) *Cache {
//     return &Cache{
//         data:    make(map[string][]byte),
//         expires: make(map[string]time.Time),
//         ttl:     ttl,
//     }
// }

// // Cache 
// func (c *Cache) getFile(path string) ([]byte, error) {
//     c.mu.RLock()
//     if data, found := c.data[path]; found {
//         if time.Now().Before(c.expires[path]) {
//             c.mu.RUnlock()
//             return data, nil
//         }
//     }
//     c.mu.RUnlock()

//     c.mu.Lock()
//     defer c.mu.Unlock()

//     if data, found := c.data[path]; found {
//         if time.Now().Before(c.expires[path]) {
//             return data, nil
//         }
//     }

//     data, err := os.ReadFile(path)
//     if err != nil {
//         return nil, err
//     }

//     c.data[path] = data
//     c.expires[path] = time.Now().Add(c.ttl)

//     return data, nil
// }

// var jwtKey = []byte("your_secret_key")

// type Claims struct {
//     Username string `json:"username"`
//     jwt.StandardClaims
// }

// // Generate JWT token
// func generateToken(username string) (string, error) {
//     expirationTime := time.Now().Add(24 * time.Hour)
//     claims := &Claims{
//         Username: username,
//         StandardClaims: jwt.StandardClaims{
//             ExpiresAt: expirationTime.Unix(),
//         },
//     }

//     token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
//     return token.SignedString(jwtKey)
// }

// // Middleware to authenticate JWT tokens
// func authenticateJWT() gin.HandlerFunc {
//     return func(c *gin.Context) {
//         tokenString := c.GetHeader("Authorization")

//         if tokenString == "" {
//             c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
//             c.Abort()
//             return
//         }

//         // Remove "Bearer " prefix if present
//         if len(tokenString) > 7 && tokenString[:7] == "Bearer " {
//             tokenString = tokenString[7:]
//         }

//         claims := &Claims{}

//         token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
//             return jwtKey, nil
//         })

//         if err != nil || !token.Valid {
//             c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
//             c.Abort()
//             return
//         }

//         c.Set("username", claims.Username)
//         c.Next()
//     }
// }

// func serveStaticFiles(c *Cache) http.HandlerFunc {
//     return func(w http.ResponseWriter, r *http.Request) {
//         if r.URL.Path == "/health" {
//             w.WriteHeader(http.StatusOK)
//             if _, err := w.Write([]byte("OK")); err != nil {
//                 http.Error(w, "Failed to write response", http.StatusInternalServerError)
//                 return
//             }
//             return
//         }

//         filePath := "." + r.URL.Path

//         data, err := c.getFile(filePath)
//         if err != nil {
//             http.Error(w, "File not found", http.StatusNotFound)
//             return
//         }

//         w.Header().Set("Content-Type", "application/octet-stream")
//         if _, err := w.Write(data); err != nil {
//             http.Error(w, "Failed to write response", http.StatusInternalServerError)
//             return
//         }
//     }
// }

// func main() {
//     cache := newCache(10 * time.Second)

//     go func() {
//         http.HandleFunc("/", serveStaticFiles(cache))
//         server := &http.Server{
//             Addr:         ":8082", 
//             ReadTimeout:  5 * time.Second,
//             WriteTimeout: 10 * time.Second,
//             IdleTimeout:  15 * time.Second,
//         }
//         fmt.Println("Starting CDN server on http://localhost:8082")
//         log.Fatal(server.ListenAndServe())
//     }()

//     r := gin.Default()

//     r.Use(func(c *gin.Context) {
//         c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
//         c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
//         c.Writer.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization")
//         if c.Request.Method == "OPTIONS" {
//             c.AbortWithStatus(http.StatusOK)
//             return
//         }
//         c.Next()
//     })

// 	r.POST("/api/login", func(c *gin.Context) {
// 		var creds struct {
// 			Username string `json:"username"`
// 			Password string `json:"password"`
// 		}
// 		if err := c.BindJSON(&creds); err != nil {
// 			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
// 			return
// 		}
	
// 		if creds.Username == "user" && creds.Password == "password" {
// 			token, err := generateToken(creds.Username)
// 			if err != nil {
// 				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
// 				return
// 			}
// 			c.JSON(http.StatusOK, gin.H{"token": token})
// 		} else {
// 			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
// 		}
// 	})

//     r.GET("/api/hello", authenticateJWT(), func(c *gin.Context) {
//         username := c.MustGet("username").(string)
//         c.JSON(http.StatusOK, gin.H{"message": "Hello " + username})
//     })

//     r.Run(":8083") 
// }


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