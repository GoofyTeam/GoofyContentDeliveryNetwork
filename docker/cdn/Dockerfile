# Étape de build
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Installation des dépendances de build
RUN apk add --no-cache git

# Copie des fichiers de dépendances
COPY app/cdn/go.mod app/cdn/go.sum ./
RUN go mod download

# Copie du reste du code source
COPY app/cdn/ .

# Compilation de l'application avec optimisations
RUN CGO_ENABLED=0 GOOS=linux go build -o main .

# Étape finale avec une image minimale
FROM alpine:latest

WORKDIR /app

# Installation des certificats CA pour les requêtes HTTPS
RUN apk --no-cache add ca-certificates

# Copie du binaire depuis l'étape de build
COPY --from=builder /app/main .
# COPY --from=builder /app/.env .

# Exposition du port
EXPOSE 8080

# Commande par défaut
CMD ["./main"]
