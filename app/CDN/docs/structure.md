# Guide de la Syntaxe Go

Ce guide explique les concepts de base de Go en utilisant des exemples de notre projet CDN.

## Table des matières

1. [Packages et Imports](#packages-et-imports)
2. [Variables et Types](#variables-et-types)
3. [Fonctions](#fonctions)
4. [Structures et Interfaces](#structures-et-interfaces)
5. [Goroutines et Channels](#goroutines-et-channels)
6. [Gestion d'Erreurs](#gestion-derreurs)

## Packages et Imports

### Déclaration de Package

```go
package main  // Déclare que ce fichier appartient au package main
```

### Imports

```go
import (
    "app/internal/cache"        // Import local (notre code)
    "github.com/sirupsen/logrus" // Import externe (dépendance)
    "net/http"                   // Import standard Go
)
```

## Variables et Types

### Déclaration de Variables

```go
// Déclaration avec initialisation
log := logrus.New()  // Type inféré automatiquement

// Déclaration explicite
var (
    backends []string = []string{"http://backend1:8080", "http://backend2:8080"}
    weights []int    = []int{1, 1}
)

// Constantes
const (
    port = ":8080"
    maxHeaderBytes = 1 << 20  // Opération bit à bit (1 MB)
)
```

### Types de Base

```go
// Nombres
var (
    port     int     = 8080        // Entier
    timeout  float64 = 10.5        // Nombre à virgule
)

// Chaînes de caractères
var url string = "http://localhost"

// Booléens
var isReady bool = true

// Durées
var timeout time.Duration = 10 * time.Second
```

### Slices (Tableaux Dynamiques)

```go
// Déclaration et initialisation
backends := []string{"http://backend1:8080", "http://backend2:8080"}

// Création avec make
handlers := make([]http.Handler, 0, 10)  // Longueur 0, capacité 10
```

## Fonctions

### Fonction Simple

```go
func main() {
    // Code de la fonction
}
```

### Fonction avec Paramètres et Retour

```go
// Fonction qui retourne plusieurs valeurs
func NewMemoryCache(size int) (*MemoryCache, error) {
    // Code
    return cache, nil
}

// Fonction anonyme (closure)
go func() {
    log.Info("Starting server on :8080")
    if err := srv.ListenAndServe(); err != nil {
        log.Fatal(err)
    }
}()
```

### Méthodes sur Types

```go
// Méthode sur le type Server
func (s *Server) ListenAndServe() error {
    // Code
}
```

## Structures et Interfaces

### Structures

```go
type Server struct {
    Addr           string
    Handler        http.Handler
    ReadTimeout    time.Duration
    WriteTimeout   time.Duration
    MaxHeaderBytes int
}

// Initialisation d'une structure
srv := &http.Server{
    Addr:           ":8080",
    Handler:        handler,
    ReadTimeout:    10 * time.Second,
    WriteTimeout:   10 * time.Second,
    MaxHeaderBytes: 1 << 20,
}
```

### Interfaces

```go
type LoadBalancer interface {
    NextBackend() *Backend
}
```

## Goroutines et Channels

### Goroutines

```go
// Lancement d'une goroutine
go func() {
    // Code exécuté en parallèle
}()
```

### Channels

```go
// Création d'un channel
quit := make(chan os.Signal, 1)

// Envoi de données dans un channel
signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

// Réception de données d'un channel
<-quit  // Attend un signal
```

## Gestion d'Erreurs

### Vérification d'Erreur Simple

```go
if err != nil {
    log.Fatal(err)
}
```

### Erreur avec Plusieurs Retours

```go
memCache, err := cache.NewMemoryCache(1000)
if err != nil {
    log.Fatal(err)
}
```

### Defer

```go
// Defer exécute la fonction juste avant de sortir du bloc actuel
defer resp.Body.Close()
defer cancel()
```

## Bonnes Pratiques

1. **Gestion des Erreurs**

   - Toujours vérifier les erreurs retournées
   - Utiliser `defer` pour le nettoyage des ressources

2. **Nommage**

   - PascalCase pour les exports (public)
   - camelCase pour le non-export (private)
   - Acronymes en majuscules (HTTP, URL, ID)

3. **Organisation du Code**

   - Un package par répertoire
   - Fichiers courts et focalisés
   - Interfaces petites et ciblées

4. **Concurrence**
   - Utiliser des goroutines pour les opérations parallèles
   - Channels pour la communication entre goroutines
   - Mutex pour la synchronisation d'accès aux ressources partagées
