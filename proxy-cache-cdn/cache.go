package main

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/go-redis/redis"
	lru "github.com/hashicorp/golang-lru"
)

var ctx = context.Background()
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

// Cache en mémoire
type Cache struct {
	store *lru.Cache
	mu    sync.Mutex
}

func NewCache(size int) *Cache {
	cache, _ := lru.New(size)
	return &Cache{store: cache}
}

// Mise en cache uniquement des réponses 200 OK
func (c *Cache) Set(key string, value []byte) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.store.Add(key, value)
}

// Récupération du cache
func (c *Cache) Get(key string) ([]byte, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	val, found := c.store.Get(key)
	if !found {
		return nil, false
	}
	return val.([]byte), true
}

// Suppression automatique des entrées les moins utilisées grâce à l’algorithme LRU
func (c *Cache) RemoveLeastUsed() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.store.Len() > 0 {
		c.store.RemoveOldest()
		log.Println("Entrée la moins utilisée supprimée du cache")
	}
}
