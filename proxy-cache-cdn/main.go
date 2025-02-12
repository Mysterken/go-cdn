package main

import (
	"fmt"
	"log"
	"net/http"

	"proxy-cache-cdn/cache"

	"proxy-cache-cdn/proxy"
)

const cacheSize = 100

var memoryCache = cache.NewCache(cacheSize)

func handler(w http.ResponseWriter, r *http.Request) {
	// Vérifier si la réponse est en cache LRU
	if val, found := memoryCache.Get(r.URL.Path); found {
		log.Println("Réponse servie depuis le cache LRU")
		w.Write(val)
		return
	}

	// Vérifier si la réponse est en cache Redis
	if val, found := cache.GetRedisCache(r.URL.Path); found {
		log.Println("Réponse servie depuis Redis")
		w.Write([]byte(val))
		memoryCache.Set(r.URL.Path, []byte(val))
		return
	}

	// Passer la requête au Reverse Proxy
	target := "http://localhost:8081"
	proxyHandler := proxy.NewReverseProxy(target)
	proxyHandler.ServeHTTP(w, r)
}

func main() {
	http.HandleFunc("/", handler)

	fmt.Println("Serveur démarré sur :9000")
	log.Fatal(http.ListenAndServe(":9000", nil))
}
