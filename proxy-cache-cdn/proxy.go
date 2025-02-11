package main

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"
)

// Liste des serveurs backend
var backends = []string{
	"http://localhost:8081",
	"http://localhost:8082",
}

// Proxy HTTP
func reverseProxy(target string) http.Handler {
	targetURL, err := url.Parse(target)
	if err != nil {
		log.Fatalf("Erreur parsing URL: %v", err)
	}

	proxy := httputil.NewSingleHostReverseProxy(targetURL)
	proxy.ErrorHandler = func(rw http.ResponseWriter, r *http.Request, err error) {
		log.Printf("Erreur proxy: %v", err)
		rw.WriteHeader(http.StatusBadGateway)
	}

	// Middleware pour journaliser les requêtes
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		start := time.Now()
		log.Printf("Requête reçue: %s %s", r.Method, r.URL.Path)
		proxy.ServeHTTP(rw, r)
		log.Printf("Temps de réponse: %s", time.Since(start))
	})
}

func main() {
	http.Handle("/", reverseProxy(backends[0]))

	log.Println("Serveur Proxy démarré sur :9000")
	if err := http.ListenAndServe(":9000", nil); err != nil {
		log.Fatal(err)
	}
}
