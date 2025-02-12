package cache

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	lru "github.com/hashicorp/golang-lru"
)

var ctx = context.Background()

// Structure du cache en mémoire (LRU)
type Cache struct {
	store *lru.Cache
	mu    sync.Mutex
}

// Création d'un cache LRU
func NewCache(size int) *Cache {
	cache, _ := lru.New(size)
	return &Cache{store: cache}
}

// Ajouter une réponse au cache mémoire
func (c *Cache) Set(key string, value []byte) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.store.Add(key, value)
}

// Récupérer une réponse depuis le cache mémoire
func (c *Cache) Get(key string) ([]byte, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	val, found := c.store.Get(key)
	if !found {
		return nil, false
	}
	return val.([]byte), true
}

// Client Redis
var redisClient = redis.NewClient(&redis.Options{
	Addr: "localhost:6379",
	DB:   0,
})

// Stocker une réponse dans Redis
func SetRedisCache(key string, value string, expiration time.Duration) {
	err := redisClient.Set(ctx, key, value, expiration).Err()
	if err != nil {
		log.Println("Erreur de mise en cache Redis:", err)
	}
}

// Récupérer une réponse depuis Redis
func GetRedisCache(key string) (string, bool) {
	val, err := redisClient.Get(ctx, key).Result()
	if err != nil {
		return "", false
	}
	return val, true
}
