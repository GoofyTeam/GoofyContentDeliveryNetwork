# ⚡ CDN Go - Réseau de Distribution de Contenu

Un **CDN (Content Delivery Network)** développé en Go, conçu pour accélérer la distribution de contenu web. Il inclut la mise en cache, l’équilibrage de charge et la surveillance des performances.

---

## 🔹 Fonctionnalités

### 🔀 Proxy HTTP

➜ Redirection dynamique des requêtes

### 🎛️ Système de Cache

✔ **Cache LRU** en mémoire  
✔ **Cache Redis** pour une meilleure scalabilité

### ⚖️ Load Balancer

✔ **Round Robin**  
✔ **Weighted Round Robin**  
✔ **Least Connections**

### 🛡️ Sécurité

✔ **Rate Limiting** (limitation de débit)  
✔ **Protection DDoS**  
✔ **Headers HTTP sécurisés**

### 📈 Monitoring & Logs

✔ **Métriques Prometheus**  
✔ **Visualisation avec Grafana**  
✔ **Logging avancé avec Logrus**

---

## 🛠 Prérequis

📌 **Outils nécessaires** :  
🔹 Docker & Docker Compose  
🔹 Go 1.23+ _(pour développement local)_

---

## 🚀 Démarrage

### 🔧 Mode Développement

Démarrer avec **hot-reload** :

```bash
docker compose -f ./docker-compose.dev.yml up -d
```

🌍 **Accès** : `http://localhost:8080`  
📊 **Métriques** : `http://localhost:8080/metrics`

### 🏭 Mode Production

Lancer une version optimisée :

```bash
docker compose -f ./docker-compose.prod.yml up -d
```

🌍 **Accès** : `http://localhost:8081`  
📊 **Métriques** : `http://localhost:8081/metrics`

### 💻 Démarrer le Frontend

```bash
cd front
npm install
npm run dev
```

### 🔗 Services Complémentaires

📊 **Grafana** : `http://localhost:3000` _(admin/admin)_  
📡 **Prometheus** : `http://localhost:9090`  
🗄️ **Redis** : `localhost:6379`

---

## 📂 Organisation du Projet

```
app/
├── back/
│   ├── internal/
│   │   ├── api/          # Gestion des routes API
│   │   ├── loadbalancer/ # Algorithmes d’équilibrage
│   │   ├── middleware/   # Sécurité & monitoring
│
├── CDN/
│   ├── config/           # Paramètres du projet
│   ├── internal/         # Code cœur du CDN
│   ├── docs/             # Documentation API
│   ├── main.go           # Point d’entrée du serveur
│
└── front/
    ├── public/          # Fichiers statiques
    ├── src/
    │   ├── assets/      # Images & icônes
    │   ├── components/  # Composants React
    │   ├── hooks/       # Hooks personnalisés
    │   ├── libs/        # Fonctions utilitaires
    │   ├── pages/       # Pages de l’app
    │   ├── routes/      # Gestion des routes
```

---

## 🔍 Détails Techniques

### 🗄️ Système de Cache

#### ⚡ Cache LRU _(en mémoire)_

📁 **Fichier** : `internal/cache/cache.go`  
✔ Gestion via `hashicorp/golang-lru`  
✔ Capacité ajustable  
✔ Cache uniquement les requêtes **GET**

#### 🛠 Gestion du Cache via API

➜ **Vider tout le cache**

```bash
curl -X POST http://localhost:8080/cache/purge
```

---

### ⚖️ Équilibrage de Charge

#### 🏗️ Algorithmes Supportés

📁 **Fichier** : `internal/loadbalancer/loadbalancer.go`

✔ **Round Robin** _(répartition cyclique)_  
✔ **Weighted Round Robin** _(distribution pondérée)_  
✔ **Least Connections** _(priorité au serveur le moins chargé)_

---

### 🌐 Endpoints API

#### 🔑 Authentification

🔹 `POST /register` ➜ Inscription  
🔹 `POST /login` ➜ Connexion

#### 📂 Gestion des Fichiers _(avec authentification)_

📥 `POST /api/files` ➜ Upload  
📤 `GET /api/files/:id` ➜ Récupération  
🗑️ `DELETE /api/files/:id` ➜ Suppression

#### 📁 Gestion des Dossiers _(avec authentification)_

📁 `POST /api/folders` ➜ Création  
📜 `GET /api/folders/:id` ➜ Liste du contenu  
🗑️ `DELETE /api/folders/:id` ➜ Suppression

#### 🔎 Monitoring

📊 `GET /metrics` ➜ Statistiques Prometheus  
💓 `GET /health` ➜ État du service  
📡 `GET /ready` ➜ Vérification de disponibilité

---

## 📊 Monitoring & Sécurité

### 📊 Métriques Disponibles

✔ **Temps de réponse** (`http_duration_seconds`)  
✔ **Total requêtes par endpoint** (`http_requests_total`)  
✔ **Taux de succès & erreurs**

### 🛡️ Mesures de Sécurité

✔ **Rate Limiting** _(100 req/s par défaut)_  
✔ **Protection XSS & Injection SQL**  
✔ **Headers Sécurisés**

- `X-Frame-Options`
- `X-Content-Type-Options`
- `X-XSS-Protection`
- `Content-Security-Policy`
- `Strict-Transport-Security`

---

## 🤝 Comment Contribuer

1️⃣ **Forkez** le repo  
2️⃣ **Créez une branche** :

```bash
git checkout -b feature/nouvelle-fonction
```

3️⃣ **Ajoutez vos changements** :

```bash
git commit -m "Ajout d'une nouvelle fonctionnalité"
```

4️⃣ **Pushez** votre code :

```bash
git push origin feature/nouvelle-fonction
```

5️⃣ **Ouvrez une Pull Request**

---

## ☁️ Déploiement sur AWS EKS

### 🔹 Pré-requis

✔ **AWS CLI** installé & configuré  
✔ **eksctl** & **kubectl** disponibles

### 🏗️ Construction de l’image Docker

```bash
docker build -t monrepo/cdn-go:latest -f docker/cdn/Dockerfile .
docker push monrepo/cdn-go:latest
```

### 🚀 Déploiement Kubernetes

```bash
eksctl create cluster --name cdn-cluster --region eu-west-3 --nodes 2
kubectl apply -f k8s/cdn-deployment.yaml
kubectl apply -f k8s/cdn-service.yaml
kubectl get pods
```

---

## 🛠 Déploiement Local avec Kubernetes

### ⚙️ Configuration

✔ Activer Kubernetes sur **Docker Desktop**  
✔ Vérifier le contexte :

```bash
kubectl config get-contexts
kubectl config use-context docker-desktop
```

### 🚀 Lancer l’application

```bash
kubectl apply -f k8s/cdn-deployment.yaml
kubectl apply -f k8s/cdn-service.yaml
kubectl get services
```

### 🔍 Vérifications

🌍 **API** : `http://localhost:80`  
📊 **Métriques** : `http://localhost:80/metrics`  
💓 **Health Check** : `http://localhost:80/health`

---

💡 **Nettoyez vos ressources après utilisation pour éviter des coûts inutiles !**

📜 _Logs & débogage_ :

```bash
kubectl logs -l app=cdn-go
```

---

✅ **CDN Go prêt à l’emploi !** 🚀
