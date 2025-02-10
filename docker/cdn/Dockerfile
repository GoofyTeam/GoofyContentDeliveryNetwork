# Étape de build
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Installation des dépendances de build
RUN apk add --no-cache git

# Copie des fichiers de dépendances
COPY app/go.mod app/go.sum ./
RUN go mod download

# Copie du code source
COPY app/CDN/ .

# Compilation de l'application avec optimisations
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o main .

# Étape finale avec une image minimale
FROM alpine:latest

WORKDIR /app

# Installation des certificats CA pour les requêtes HTTPS
RUN apk --no-cache add ca-certificates

# Copie du binaire depuis l'étape de build
COPY --from=builder /app/main .

# Exposition du port
EXPOSE 8080

# Commande de démarrage
CMD ["./main"]
