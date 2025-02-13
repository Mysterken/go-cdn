# Utiliser l'image officielle Go comme environnement de construction
FROM golang:1.23-alpine AS builder

# Définir le répertoire de travail dans le conteneur
WORKDIR /app

# Copier les fichiers de module Go et installer les dépendances
COPY go.mod go.sum ./
RUN go mod tidy

# Copier les fichiers de l'application
COPY . .

# Construire l'application Go
RUN go build -o go-cdn main.go

# Utiliser une image de base minimale pour le conteneur final
FROM alpine:3.21

# Créer un nouvel utilisateur et groupe
RUN addgroup -S gocdn-group && adduser -S gocdn-user -G gocdn-group

# Définir le répertoire de travail dans le conteneur
WORKDIR /app

# Copier le binaire construit à partir de l'étape de construction
COPY --from=builder /app/go-cdn /app/go-cdn

# Copier les fichiers statiques depuis l'hôte (assurez-vous que /static existe)
COPY static /app/static

# Changer le propriétaire des fichiers de l'application pour le nouvel utilisateur
RUN chown -R gocdn-user:gocdn-group /app

# Passer à l'utilisateur gocdn
USER gocdn-user

# Exposer le port 8080
EXPOSE 8080

# Ajouter une instruction HEALTHCHECK
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
  CMD ["wget", "--spider", "http://localhost:8080/health"]

# Démarrer l'application
CMD ["/app/go-cdn"]
