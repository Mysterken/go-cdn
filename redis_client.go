package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)


var ctx = context.Background()

// Initialisation du client Redis
func NewRedisClient() *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379", // Adresse du serveur Redis
		Password: "",               // Pas de mot de passe par défaut
		DB:       0,                // Base de données Redis (par défaut)
	})

	// Test de connexion
	_, err := client.Ping(ctx).Result()
	if err != nil {
		log.Fatalf("Erreur de connexion à Redis : %v", err)
	}
	fmt.Println("✅ Connecté à Redis")
	return client
}

// Stocker une valeur dans Redis
func SetCache(client *redis.Client, key string, value string, expiration time.Duration) {
	err := client.Set(ctx, key, value, expiration).Err()
	if err != nil {
		log.Printf("Erreur lors de l'enregistrement dans Redis : %v", err)
	}
}

// Récupérer une valeur depuis Redis
func GetCache(client *redis.Client, key string) string {
	val, err := client.Get(ctx, key).Result()
	if err == redis.Nil {
		log.Printf("Clé %s non trouvée dans Redis", key)
		return ""
	} else if err != nil {
		log.Printf("Erreur de récupération depuis Redis : %v", err)
		return ""
	}
	return val
}

func main() {
	client := NewRedisClient()

	// Stocker et récupérer un test
	SetCache(client, "cdn:test", "Hello Redis!", 10*time.Second)
	val := GetCache(client, "cdn:test")
	fmt.Println("Valeur récupérée depuis Redis :", val)
}
