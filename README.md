# CDN Go – Projet de Réseau de Diffusion de Contenu

Ce projet, développé en Go, met en place un Content Delivery Network (CDN) afin d’optimiser la distribution de contenu web. Il intègre des mécanismes avancés de mise en cache, de répartition de charge et de monitoring.

## 🚀 Fonctionnalités

- **Proxy HTTP** : Redirection intelligente des requêtes
- **Mécanisme de Cache** :
  - Cache LRU en mémoire
  - Intégration de Redis pour un cache distribué
- **Répartition de Charge** :
  - Round Robin
  - Weighted Round Robin
  - Least Connections
- **Sécurité** :
  - Limitation du débit (Rate Limiting)
  - Protection contre les attaques DDoS
  - Application de headers de sécurité HTTP
- **Monitoring** :
  - Collecte de métriques via Prometheus
  - Visualisation avec Grafana
  - Logging structuré grâce à Logrus

## 🛠 Prérequis

- Docker
- Docker Compose
- Go 1.23 ou supérieur (pour le développement local)

## 🚦 Démarrage

### 1. Mode Développement

Lancer l’application en mode développement avec hot-reload :

```bash
docker compose -f docker-compose.dev.yml up
```

- Accessible via [http://localhost:8080](http://localhost:8080)
- Les métriques sont disponibles sur [http://localhost:8080/metrics](http://localhost:8080/metrics)

### 2. Mode Production

Démarrer en mode production :

```bash
docker compose -f docker-compose.prod.yml up
```

- Optimisé pour un environnement de production
- Accessible via [http://localhost:8081](http://localhost:8081)
- Les métriques se trouvent sur [http://localhost:8081/metrics](http://localhost:8081/metrics)

### 3. Services Complémentaires

- **Grafana** : [http://localhost:3000](http://localhost:3000) (identifiants par défaut : admin/admin)
- **Prometheus** : [http://localhost:9090](http://localhost:9090)
- **Redis** : Accessible sur localhost:6379

## 🏗 Organisation du Projet

```
app/
├── internal/
│   ├── cache/          # Gestion du cache (implémentation LRU et intégration Redis)
│   ├── loadbalancer/   # Algorithmes de répartition de charge
│   └── middleware/     # Middlewares pour la sécurité et le monitoring
├── pkg/
│   └── config/         # Fichiers de configuration de l’application
└── main.go             # Point d’entrée de l’application
```

## 🔍 Fonctionnement en Détail

### 1. Système de Cache

- **Cache LRU** (`internal/cache/cache.go`) :
  - Respecte l’interface `Cache`
  - S’appuie sur la librairie `hashicorp/golang-lru` pour la gestion en mémoire
  - Taille du cache configurable
  - Cible uniquement les requêtes GET
  - Durée de vie (TTL) des entrées paramétrable

- **Endpoints de Gestion du Cache** :
  - `POST /cache/purge` : Permet de vider l’intégralité du cache  
    Exemple d’utilisation :
    ```bash
    curl -X POST http://localhost:8080/cache/purge
    ```

### 2. Load Balancer

- **Implémentations** (voir `internal/loadbalancer/loadbalancer.go`) :
  - **RoundRobin** : Distribution cyclique des requêtes
  - **WeightedRoundRobin** : Distribution pondérée en fonction des capacités des serveurs
  - **LeastConnections** : Acheminement vers le serveur avec le moins de connexions actives

### 3. Endpoints API

#### Backend Service (port 8080)

- **Authentification** :
  - `POST /register` : Inscription d’un nouvel utilisateur
  - `POST /login` : Connexion d’un utilisateur

- **Gestion des Fichiers** *(authentification requise)* :
  - `POST /api/files` : Upload d’un fichier
  - `GET /api/files/:id` : Récupération d’un fichier
  - `DELETE /api/files/:id` : Suppression d’un fichier

- **Gestion des Dossiers** *(authentification requise)* :
  - `POST /api/folders` : Création d’un dossier
  - `GET /api/folders/:id` : Affichage du contenu d’un dossier
  - `DELETE /api/folders/:id` : Suppression d’un dossier

- **Health Check** :
  - `GET /health` : Vérification de l’état du service

#### CDN Service (port 8080)

- **Cache** :
  - `POST /cache/purge` : Effacement du cache
  - *Note* : Seules les requêtes GET sont mises en cache

- **Monitoring** :
  - `GET /metrics` : Exposition des métriques Prometheus
  - `GET /health` : État de santé du CDN
  - `GET /ready` : Vérification de la disponibilité

### 4. Monitoring

- **Métriques Collectées** :
  - Temps de réponse des requêtes
  - Nombre de requêtes par endpoint
  - Taux de réussite vs. échec
  - Utilisation du cache

- **Visualisation** : Les données sont exploitées dans Grafana via Prometheus

### 5. Application Principale

Le fichier `main.go` orchestre l’ensemble des composants en :
1. Initialisant le logger et le cache
2. Configurant le load balancer
3. Déployant les middlewares pour la sécurité et le monitoring
4. Démarrant le serveur HTTP avec une gestion gracieuse de l’arrêt

## 📊 Monitoring

### Métriques Disponibles :

- `http_duration_seconds` : Mesure du temps de réponse des requêtes
- `http_requests_total` : Compte total des requêtes par endpoint

Les visualisations se font via Grafana, en s’appuyant sur Prometheus.

## 🔒 Sécurité

- **Rate Limiting** : Limitation par défaut à 100 requêtes par seconde
- **Headers de Sécurité** :
  - `X-Frame-Options`
  - `X-Content-Type-Options`
  - `X-XSS-Protection`
  - `Content-Security-Policy`
  - `Strict-Transport-Security`

## 🤝 Contribution

Pour contribuer :

1. Forkez le projet
2. Créez votre branche de travail (par exemple : `git checkout -b feature/amazing-feature`)
3. Effectuez vos commits (`git commit -m 'Ajout d’une fonctionnalité géniale'`)
4. Poussez votre branche (`git push origin feature/amazing-feature`)
5. Ouvrez une Pull Request

## 🚀 Déploiement sur AWS EKS

### Prérequis AWS

- Un compte AWS avec les droits nécessaires
- AWS CLI configuré
- `eksctl` installé
- `kubectl` installé

### 1. Construction de l’Image Docker

```bash
# Construction de l’image Docker
docker build -t adr181100/goofy-cdn:latest -f docker/cdn/Dockerfile .

# Envoi de l’image sur Docker Hub
docker push adr181100/goofy-cdn:latest
```

### 2. Déploiement sur EKS

#### Création du Cluster

```bash
eksctl create cluster \
  --name goofy-cdn-cluster \
  --region eu-west-3 \
  --nodegroup-name goofy-cdn-workers \
  --node-type t3.small \
  --nodes 2 \
  --nodes-min 1 \
  --nodes-max 3
```

#### Déploiement de l’Application

```bash
# Déploiement via Kubernetes
kubectl apply -f k8s/cdn-deployment.yaml
kubectl apply -f k8s/cdn-service.yaml

# Vérification du déploiement
kubectl get pods
kubectl get services
```

### 3. Gestion des Ressources

#### Vérification

```bash
# Afficher les nœuds du cluster
kubectl get nodes

# Lister tous les pods
kubectl get pods --all-namespaces

# Afficher les logs des pods associés
kubectl logs -l app=goofy-cdn
```

#### Nettoyage

```bash
# Supprimer le nodegroup
eksctl delete nodegroup --cluster goofy-cdn-cluster --name goofy-cdn-workers

# Supprimer le cluster complet (pour éviter des coûts supplémentaires)
eksctl delete cluster --name goofy-cdn-cluster
```

### 4. Surveillance des Coûts AWS

- **Cluster EKS** : environ 0,10 $ par heure
- **Nœuds EC2 (t3.small)** : environ 0,023 $ par heure par nœud
- **Load Balancer** : environ 0,025 $ par heure
- **Volumes EBS et ENI** : coûts variables selon l’utilisation

⚠️ **Important** : Veillez à supprimer l’ensemble des ressources après usage pour éviter des frais inutiles.

### 5. Dépannage Courant

#### Problèmes de CNI

```bash
# Réinstaller le CNI Amazon VPC
kubectl apply -f https://raw.githubusercontent.com/aws/amazon-vpc-cni-k8s/v1.12.6/config/master/aws-k8s-cni.yaml

# Redémarrer les pods du CNI
kubectl delete pods -n kube-system -l k8s-app=aws-node
```

#### Problèmes de Permissions

Assurez-vous que le rôle IAM possède bien les politiques suivantes :

- AmazonEKSClusterPolicy
- AmazonEKSServicePolicy
- AmazonEKSVPCResourceController
- AmazonEKS_CNI_Policy

---

## 🖥 Déploiement Local avec Docker Desktop

### Prérequis

- Docker Desktop installé
- Kubernetes activé dans Docker Desktop (via kubeadm)
- `kubectl` installé (ex. : `brew install kubectl`)

### 1. Configuration de Kubernetes dans Docker Desktop

1. Ouvrez Docker Desktop  
2. Rendez-vous dans **Settings > Kubernetes**  
3. Cochez **Enable Kubernetes**  
4. Sélectionnez **kubeadm** comme méthode de provisionnement  
5. Cliquez sur **Apply & Restart**

### 2. Construction de l’Image

```bash
# Construire l’image localement
docker build -t goofy-cdn:local -f docker/cdn/Dockerfile .
```

### 3. Déploiement sur Kubernetes Local

1. **Vérifier le Contexte de kubectl** :

    ```bash
    # Afficher les contextes disponibles
    kubectl config get-contexts

    # Utiliser le contexte Docker Desktop si nécessaire
    kubectl config use-context docker-desktop
    ```

2. **Déployer l’Application** :

    ```bash
    # Appliquer les fichiers de configuration Kubernetes
    kubectl apply -f k8s/cdn-deployment.yaml
    kubectl apply -f k8s/cdn-service.yaml

    # Vérifier l’état des pods et services
    kubectl get pods
    kubectl get services
    ```

### 4. Accès à l’Application

L’application est accessible aux adresses suivantes :

- **URL Principale** : [http://localhost:80](http://localhost:80)
- **Métriques** : [http://localhost:80/metrics](http://localhost:80/metrics)
- **Health Check** : [http://localhost:80/health](http://localhost:80/health)
- **Readiness** : [http://localhost:80/ready](http://localhost:80/ready)

### 5. Commandes Utiles

```bash
# Afficher les logs de l’application
kubectl logs -l app=goofy-cdn

# Obtenir les détails d’un pod
kubectl describe pod -l app=goofy-cdn

# Redémarrer les pods (après modification du code)
kubectl delete pod -l app=goofy-cdn

# Supprimer le déploiement
kubectl delete -f k8s/cdn-deployment.yaml
kubectl delete -f k8s/cdn-service.yaml
```

### 6. Dépannage

#### Pods en CrashLoopBackOff ou Erreur

```bash
# Consulter les logs du pod
kubectl logs -l app=goofy-cdn

# Afficher les détails et événements du pod
kubectl describe pod -l app=goofy-cdn
```

#### Service Inaccessible

1. Vérifier que le service est bien créé :
    ```bash
    kubectl get services
    ```

2. S’assurer que le pod est en état Ready :
    ```bash
    kubectl get pods -l app=goofy-cdn
    ```

3. Visualiser les endpoints associés :
    ```bash
    kubectl get endpoints goofy-cdn-service
    ```

#### Problèmes d’Image

Si l’image n’est pas trouvée, vérifiez que :
1. L’image est bien construite localement :
    ```bash
    docker images | grep goofy-cdn
    ```
2. Le fichier de déploiement utilise le bon nom d’image : `image: goofy-cdn:local`
