package routes

import (
	"fmt"
	"io"
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

// fileManager gère la route /files pour créer, modifier ou supprimer des fichiers.
func FileManager(w http.ResponseWriter, r *http.Request) {
	// Définir le dossier de stockage
	filesDir := "files"

	// Vérifier l'existence du dossier, sinon le créer
	if _, err := os.Stat(filesDir); os.IsNotExist(err) {
		if err := os.Mkdir(filesDir, 0755); err != nil {
			http.Error(w, "Erreur lors de la création du dossier files: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}

	switch r.Method {
	case http.MethodPost:
		// Créer un nouveau fichier
		// On s'attend à recevoir un fichier dans le champ "file" du formulaire
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
		// Modifier (mettre à jour) un fichier existant
		// Le nom du fichier doit être précisé en query paramètre "filename"
		filename := r.URL.Query().Get("filename")
		if filename == "" {
			http.Error(w, "Le paramètre 'filename' est requis pour modifier un fichier", http.StatusBadRequest)
			return
		}

		filePath := filepath.Join(filesDir, filename)
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			http.Error(w, "Fichier non trouvé", http.StatusNotFound)
			return
		}

		// Ouvrir le fichier en écriture et tronquer son contenu
		f, err := os.OpenFile(filePath, os.O_WRONLY|os.O_TRUNC, 0644)
		if err != nil {
			http.Error(w, "Erreur lors de l'ouverture du fichier pour modification: "+err.Error(), http.StatusInternalServerError)
			return
		}
		defer f.Close()

		// Copier le nouveau contenu depuis le corps de la requête
		if _, err := io.Copy(f, r.Body); err != nil {
			http.Error(w, "Erreur lors de la mise à jour du fichier: "+err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "Fichier %s modifié avec succès", filename)

	case http.MethodDelete:
		// Supprimer un fichier
		// Le nom du fichier doit être précisé en query paramètre "filename"
		filename := r.URL.Query().Get("filename")
		if filename == "" {
			http.Error(w, "Le paramètre 'filename' est requis pour supprimer un fichier", http.StatusBadRequest)
			return
		}

		filePath := filepath.Join(filesDir, filename)
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

	default:
		// Méthode HTTP non autorisée
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
	}
}
