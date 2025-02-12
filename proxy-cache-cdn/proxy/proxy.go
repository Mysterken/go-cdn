package proxy

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

// Fonction pour créer un Reverse Proxy
func NewReverseProxy(target string) http.Handler {
	targetURL, err := url.Parse(target)
	if err != nil {
		log.Fatalf("Erreur parsing URL: %v", err)
	}

	proxy := httputil.NewSingleHostReverseProxy(targetURL)
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		log.Printf("Erreur proxy: %v", err)
		w.WriteHeader(http.StatusBadGateway)
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		log.Printf("Requête reçue: %s %s", r.Method, r.URL.Path)
		proxy.ServeHTTP(w, r)
		log.Printf("Temps de réponse: %s", time.Since(start))
	})
}
