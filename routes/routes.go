package routes

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

// DownloadCat permet de télécharger le fichier cat.jpg situé dans le dossier static.
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

// DownloadImage permet de télécharger dynamiquement une image depuis le dossier static.
func DownloadImage(w http.ResponseWriter, r *http.Request) {
	prefix := "/download/"
	if len(r.URL.Path) <= len(prefix) {
		http.Error(w, "Nom de l'image non fourni", http.StatusBadRequest)
		return
	}

	// Extraire et sécuriser le nom de l'image
	imageName := filepath.Base(r.URL.Path[len(prefix):])
	filePath := filepath.Join("static", imageName)

	// Vérifier que le fichier existe
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		log.Println("Fichier non trouvé :", filePath)
		http.Error(w, "Fichier non trouvé", http.StatusNotFound)
		return
	}

	// Définir les en-têtes pour forcer le téléchargement
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", "attachment; filename=\""+imageName+"\"")

	http.ServeFile(w, r, filePath)
}

// fileManager gère la création, la modification, la suppression et la liste des fichiers dans le dossier "files".
func FileManager(w http.ResponseWriter, r *http.Request) {
	filesDir := "files"
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	// Répondre tout de suite aux requêtes OPTIONS (pré-vol)
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Créer le dossier s'il n'existe pas
	if _, err := os.Stat(filesDir); os.IsNotExist(err) {
		if err := os.Mkdir(filesDir, 0755); err != nil {
			http.Error(w, "Erreur lors de la création du dossier files: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}

	switch r.Method {
	case http.MethodPost:
		// Création d'un nouveau fichier
		file, handler, err := r.FormFile("file")
		if err != nil {
			http.Error(w, "Erreur lors de la récupération du fichier: "+err.Error(), http.StatusBadRequest)
			return
		}
		defer file.Close()

		filePath := filepath.Join(filesDir, handler.Filename)
		dst, err := os.Create(filePath)
		if err != nil {
			http.Error(w, "Erreur lors de la création du fichier: "+err.Error(), http.StatusInternalServerError)
			return
		}
		defer dst.Close()

		if _, err := io.Copy(dst, file); err != nil {
			http.Error(w, "Erreur lors de la sauvegarde du fichier: "+err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
		fmt.Fprintf(w, "Fichier %s créé avec succès", handler.Filename)

	case http.MethodPut:
		// Modification d'un fichier existant
		filename := r.URL.Query().Get("filename")
		if filename == "" {
			http.Error(w, "Le paramètre 'filename' est requis pour modifier un fichier", http.StatusBadRequest)
			return
		}

		filePath := filepath.Join(filesDir, filename)
		fmt.Println("PUT: Recherche du fichier à modifier :", filePath)
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			http.Error(w, "Fichier non trouvé", http.StatusNotFound)
			return
		}

		// Ouvrir le fichier en écriture en tronquant son contenu
		f, err := os.OpenFile(filePath, os.O_WRONLY|os.O_TRUNC, 0644)
		if err != nil {
			http.Error(w, "Erreur lors de l'ouverture du fichier pour modification: "+err.Error(), http.StatusInternalServerError)
			return
		}
		defer f.Close()

		if _, err := io.Copy(f, r.Body); err != nil {
			http.Error(w, "Erreur lors de la modification du fichier: "+err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "Fichier %s modifié avec succès", filename)

	case http.MethodDelete:
		// Suppression d'un fichier existant
		filename := r.URL.Query().Get("filename")
		if filename == "" {
			http.Error(w, "Le paramètre 'filename' est requis pour supprimer un fichier", http.StatusBadRequest)
			return
		}

		filePath := filepath.Join(filesDir, filename)
		fmt.Println("DELETE: Recherche du fichier à supprimer :", filePath)
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			http.Error(w, "Fichier non trouvé", http.StatusNotFound)
			return
		}

		if err := os.Remove(filePath); err != nil {
			http.Error(w, "Erreur lors de la suppression du fichier: "+err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "Fichier %s supprimé avec succès", filename)

	case http.MethodGet:
		// Lister les fichiers présents dans le dossier "files"
		files, err := os.ReadDir(filesDir)
		if err != nil {
			http.Error(w, "Erreur lors de la lecture du dossier: "+err.Error(), http.StatusInternalServerError)
			return
		}
		for _, file := range files {
			fmt.Fprintln(w, file.Name())
		}

	default:
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
	}
}
