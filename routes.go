// routes/routes.go
package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
)

// Exporter la fonction en renommant downloadCat en DownloadCat
func DownloadCat(w http.ResponseWriter, r *http.Request) {
	filePath := filepath.Join("static", "cat.jpg")

	// Vérifier que le fichier existe
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}

	// Définir les en-têtes pour forcer le téléchargement
	w.Header().Set("Content-Type", "image/jpeg")
	w.Header().Set("Content-Disposition", "attachment; filename=\"cat.jpg\"")

	// Envoyer le fichier
	http.ServeFile(w, r, filePath)
}

func DownloadImage(w http.ResponseWriter, r *http.Request) {
	// Définir le préfixe de la route
	prefix := "/download/"
	// Vérifier que l'URL contient bien le nom de l'image
	if len(r.URL.Path) <= len(prefix) {
		http.Error(w, "Nom de l'image non fourni", http.StatusBadRequest)
		return
	}

	// Extraire le nom de l'image depuis l'URL
	imageName := r.URL.Path[len(prefix):]
	// Sécuriser le nom de l'image en conservant uniquement le nom de base (évite les attaques par chemin)
	imageName = filepath.Base(imageName)

	// Construire le chemin complet vers le fichier dans le dossier "static"
	filePath := filepath.Join("static", imageName)

	// Vérifier que le fichier existe
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		log.Println("Fichier non trouvé :", filePath)
		http.Error(w, "Fichier non trouvé", http.StatusNotFound)
		return
	}

	// Définir les en-têtes pour forcer le téléchargement
	// Ici, on utilise "application/octet-stream" pour indiquer un téléchargement générique.
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", "attachment; filename=\""+imageName+"\"")

	// Servir le fichier au client
	http.ServeFile(w, r, filePath)
}
