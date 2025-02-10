FROM golang:1.23-alpine AS builder

WORKDIR /app

# Installation des dépendances système
RUN apk add --no-cache git

# Copie des fichiers de dépendances
COPY app/back/go.mod app/back/go.sum ./

# Téléchargement des dépendances
RUN go mod download

# Copie du reste du code source
COPY app/back/ .

# Construction de l'application
RUN CGO_ENABLED=0 GOOS=linux go build -o main .

# Image finale légère
FROM alpine:latest

WORKDIR /app

# Copie du binaire depuis l'étape de build
COPY --from=builder /app/main .
COPY --from=builder /app/.env .

# Exposition du port
EXPOSE 8080

# Commande de démarrage
CMD ["./main"]
