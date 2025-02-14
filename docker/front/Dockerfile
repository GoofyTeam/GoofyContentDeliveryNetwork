# Étape de build
FROM --platform=linux/amd64 node:20 AS builder

WORKDIR /app

# Copie des fichiers de dépendances
COPY app/front/package*.json ./

# Installation des dépendances
RUN npm install

# Copie du reste du code source
COPY app/front/ ./

# Build de l'application
RUN npm run build

# Étape de production avec Nginx
FROM nginx:alpine

# Copie de la configuration nginx
COPY docker/front/nginx.conf /etc/nginx/conf.d/default.conf

# Copie des fichiers buildés
COPY --from=builder /app/dist /usr/share/nginx/html

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]