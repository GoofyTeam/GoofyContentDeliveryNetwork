# CDN Go - Projet de Content Delivery Network

Ce projet implémente un Content Delivery Network (CDN) en Go, conçu pour optimiser la distribution de contenu web avec des fonctionnalités avancées de mise en cache, de répartition de charge et de monitoring.

## 🚀 Fonctionnalités

- **Proxy HTTP** : Redirection intelligente des requêtes
- **Système de Cache** :
  - Cache LRU en mémoire
  - Support Redis pour le cache distribué
- **Load Balancing** :
  - Round Robin
  - Weighted Round Robin
  - Least Connections
- **Sécurité** :
  - Rate Limiting
  - Protection DDoS
  - Headers de sécurité HTTP
- **Monitoring** :
  - Métriques Prometheus
  - Visualisation Grafana
  - Logging structuré avec Logrus

## 🛠 Prérequis

- Docker
- Docker Compose
- Go 1.23+ (pour le développement local)

## 🚦 Démarrage

1. **Mode Développement** :

```bash
docker compose up app-dev
```

- Hot-reload activé
- Accessible sur http://localhost:8080
- Métriques sur http://localhost:8080/metrics

2. **Mode Production** :

```bash
docker compose up app-prod
```

- Optimisé pour la production
- Accessible sur http://localhost:8081
- Métriques sur http://localhost:8081/metrics

3. **Services additionnels** :

- Grafana : http://localhost:3000 (admin/admin)
- Prometheus : http://localhost:9090
- Redis : localhost:6379

## 🏗 Structure du Projet

```
app/
├── internal/
│   ├── cache/          # Implémentation du cache (LRU, Redis)
│   ├── loadbalancer/   # Algorithmes de load balancing
│   └── middleware/     # Middlewares (sécurité, métriques)
├── pkg/
│   └── config/         # Configuration de l'application
└── main.go            # Point d'entrée de l'application
```

## 🔍 Fonctionnement Détaillé

### 1. Système de Cache

- **Cache LRU** (`internal/cache/cache.go`) :
  - Implémente l'interface `Cache`
  - Utilise `hashicorp/golang-lru` pour la gestion du cache en mémoire
  - Limite configurable de la taille du cache
  - Cache uniquement les requêtes GET
  - TTL configurable pour les entrées du cache

- **Endpoints de Gestion du Cache** :
  - `POST /cache/purge` : Vide complètement le cache
    ```bash
    # Exemple d'utilisation
    curl -X POST http://localhost:8080/cache/purge
    ```

### 2. Load Balancer

- **Implémentations** (`internal/loadbalancer/loadbalancer.go`) :
  - `RoundRobin` : Distribution cyclique des requêtes
  - `WeightedRoundRobin` : Distribution pondérée selon la capacité des serveurs
  - `LeastConnections` : Envoi vers le serveur le moins chargé

### 3. Endpoints API

#### Backend Service (port 8080)
- **Authentification** :
  - `POST /register` : Inscription d'un nouvel utilisateur
  - `POST /login` : Connexion utilisateur

- **Gestion des Fichiers** (requiert authentification) :
  - `POST /api/files` : Upload d'un fichier
  - `GET /api/files/:id` : Récupération d'un fichier
  - `DELETE /api/files/:id` : Suppression d'un fichier

- **Gestion des Dossiers** (requiert authentification) :
  - `POST /api/folders` : Création d'un dossier
  - `GET /api/folders/:id` : Liste du contenu d'un dossier
  - `DELETE /api/folders/:id` : Suppression d'un dossier

- **Health Check** :
  - `GET /health` : Vérification de l'état du service

#### CDN Service (port 8080)
- **Cache** :
  - `POST /cache/purge` : Vide le cache
  - Note : Seules les requêtes GET sont mises en cache

- **Monitoring** :
  - `GET /metrics` : Métriques Prometheus
  - `GET /health` : État du CDN
  - `GET /ready` : Vérification de disponibilité

### 4. Monitoring

- **Métriques** :
  - Temps de réponse des requêtes
  - Nombre de requêtes par endpoint
  - Taux de succès/erreur
  - Utilisation du cache

- **Visualisation dans Grafana** via Prometheus

### 5. Application Principale

Le fichier `main.go` orchestre tous ces composants :

1. Initialise le logger et le cache
2. Configure le load balancer
3. Met en place les middlewares de sécurité et monitoring
4. Démarre le serveur HTTP avec gestion gracieuse de l'arrêt

## 📊 Monitoring

### Métriques disponibles :

- `http_duration_seconds` : Temps de réponse des requêtes
- `http_requests_total` : Nombre total de requêtes par endpoint
- Visualisation dans Grafana via Prometheus

## 🔒 Sécurité

- Rate limiting : 100 requêtes/seconde par défaut
- Headers de sécurité :
  - `X-Frame-Options`
  - `X-Content-Type-Options`
  - `X-XSS-Protection`
  - `Content-Security-Policy`
  - `Strict-Transport-Security`

## 🤝 Contribution

1. Fork le projet
2. Créez votre branche (`git checkout -b feature/amazing-feature`)
3. Committez vos changements (`git commit -m 'Add amazing feature'`)
4. Push vers la branche (`git push origin feature/amazing-feature`)
5. Ouvrez une Pull Request

## 🚀 Déploiement sur AWS EKS

### Prérequis AWS

- Un compte AWS avec les droits nécessaires
- AWS CLI configuré
- `eksctl` installé
- `kubectl` installé

### 1. Construction de l'Image Docker

```bash
# Construction de l'image
docker build -t misterzapp/goofy-cdn:latest -f docker/cdn/Dockerfile .

# Push vers Docker Hub
docker push misterzapp/goofy-cdn:latest
```

### 2. Déploiement sur EKS

#### Création du Cluster

```bash
# Création du cluster EKS
eksctl create cluster \
  --name goofy-cdn-cluster \
  --region eu-west-3 \
  --nodegroup-name goofy-cdn-workers \
  --node-type t3.small \
  --nodes 2 \
  --nodes-min 1 \
  --nodes-max 3
```

#### Déploiement de l'Application

```bash
# Déployer l'application
kubectl apply -f k8s/cdn-deployment.yaml
kubectl apply -f k8s/cdn-service.yaml

# Vérifier le déploiement
kubectl get pods
kubectl get services
```

### 3. Gestion des Ressources

#### Vérification des Ressources

```bash
# Lister les nœuds
kubectl get nodes

# Lister les pods
kubectl get pods --all-namespaces

# Voir les logs
kubectl logs -l app=goofy-cdn
```

#### Nettoyage des Ressources

```bash
# Supprimer le nodegroup
eksctl delete nodegroup --cluster goofy-cdn-cluster --name goofy-cdn-workers

# Supprimer le cluster complet (arrête toute facturation)
eksctl delete cluster --name goofy-cdn-cluster
```

### 4. Coûts AWS à Surveiller

- Cluster EKS : ~$0.10 par heure
- Nœuds EC2 (t3.small) : ~$0.023 par heure par nœud
- Load Balancer : ~$0.025 par heure
- Volumes EBS et ENI : coûts variables selon l'utilisation

⚠️ **Important** : Pensez à supprimer toutes les ressources après utilisation pour éviter des coûts inutiles.

### 5. Troubleshooting Courant

#### Problèmes de CNI ( a voir car problème pour l'instant)

Si les pods restent en état "ContainerCreating" :

```bash
# Réinstaller le CNI Amazon VPC
kubectl apply -f https://raw.githubusercontent.com/aws/amazon-vpc-cni-k8s/v1.12.6/config/master/aws-k8s-cni.yaml

# Redémarrer les pods CNI
kubectl delete pods -n kube-system -l k8s-app=aws-node
```

#### Problèmes de Permissions

Vérifier que le rôle IAM a les bonnes politiques :

- AmazonEKSClusterPolicy
- AmazonEKSServicePolicy
- AmazonEKSVPCResourceController
- AmazonEKS_CNI_Policy

## 🖥 Déploiement Local avec Docker Desktop

### Prérequis
- Docker Desktop installé
- Kubernetes activé dans Docker Desktop (avec kubeadm)
- kubectl installé (`brew install kubectl`)

### 1. Configuration de Kubernetes dans Docker Desktop
1. Ouvrir Docker Desktop
2. Aller dans Settings > Kubernetes
3. Sélectionner "Enable Kubernetes"
4. Choisir "kubeadm" comme méthode de provisionnement
5. Cliquer sur "Apply & Restart"

### 2. Construction de l'Image
```bash
# Construire l'image localement
docker build -t goofy-cdn:local -f docker/cdn/Dockerfile .
```

### 3. Déploiement sur Kubernetes Local

1. **Vérifier que kubectl utilise le bon contexte** :
```bash
# Voir les contextes disponibles
kubectl config get-contexts

# Passer au contexte Docker Desktop si nécessaire
kubectl config use-context docker-desktop
```

2. **Déployer l'application** :
```bash
# Appliquer les configurations
kubectl apply -f k8s/cdn-deployment.yaml
kubectl apply -f k8s/cdn-service.yaml

# Vérifier le déploiement
kubectl get pods
kubectl get services
```

### 4. Accès à l'Application

L'application est accessible via les endpoints suivants :
- **URL Principale** : `http://localhost:80`
- **Métriques** : `http://localhost:80/metrics`
- **Health Check** : `http://localhost:80/health`
- **Readiness** : `http://localhost:80/ready`

### 5. Commandes Utiles

```bash
# Voir les logs de l'application
kubectl logs -l app=goofy-cdn

# Voir les détails du pod
kubectl describe pod -l app=goofy-cdn

# Redémarrer le déploiement (après modification du code)
kubectl delete pod -l app=goofy-cdn

# Supprimer le déploiement
kubectl delete -f k8s/cdn-deployment.yaml
kubectl delete -f k8s/cdn-service.yaml
```

### 6. Troubleshooting

#### Pod en CrashLoopBackOff ou Error
```bash
# Voir les logs du pod
kubectl logs -l app=goofy-cdn

# Voir les détails et événements du pod
kubectl describe pod -l app=goofy-cdn
```

#### Service inaccessible
1. Vérifier que le service est bien créé :
```bash
kubectl get services
```

2. Vérifier que le pod est Ready :
```bash
kubectl get pods -l app=goofy-cdn
```

3. Voir les endpoints :
```bash
kubectl get endpoints goofy-cdn-service
```

#### Problèmes d'image
Si l'image n'est pas trouvée, assurez-vous que :
1. L'image est bien construite localement : `docker images | grep goofy-cdn`
2. Le fichier deployment.yaml utilise le bon nom d'image : `image: goofy-cdn:local`
