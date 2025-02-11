// routes/routes.go
package main

import (
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
