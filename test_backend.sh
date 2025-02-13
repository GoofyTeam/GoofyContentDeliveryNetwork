#!/bin/bash

BACKEND_URL="http://localhost:8080"  # Ajustez le port selon votre configuration

echo "Démarrage des tests du backend..."

# Test de téléchargement
echo "Test de téléchargement..."
curl -s "$BACKEND_URL/test/download/small" -o small.bin &
curl -s "$BACKEND_URL/test/download/medium" -o medium.bin &
curl -s "$BACKEND_URL/test/download/large" -o large.bin &

# Test d'upload
echo "Test d'upload..."
dd if=/dev/urandom of=test_10mb.bin bs=1M count=10 2>/dev/null
curl -s -F "file=@test_10mb.bin" "$BACKEND_URL/test/upload" > /dev/null &

# Test de latence
echo "Test des patterns de latence..."
for pattern in spike wave random; do
    for i in {1..5}; do
        curl -s "$BACKEND_URL/test/latency/$pattern" > /dev/null &
    done
done

# Test de streaming
echo "Test de streaming..."
curl -s "$BACKEND_URL/test/stream/10" > /dev/null &

# Test de compression
echo "Test de compression..."
curl -s -H "Accept-Encoding: gzip" "$BACKEND_URL/test/compression" > /dev/null &

# Test d'erreurs
echo "Test des erreurs..."
for error in timeout memory cpu; do
    curl -s "$BACKEND_URL/test/error/$error" > /dev/null &
done

# Test de cache avec contenu statique
echo "Test de cache avec contenu statique..."
for i in {1..10}; do
    # Appeler la même URL plusieurs fois pour tester le cache
    curl -s "$BACKEND_URL/test/cache/static/1" > /dev/null
    curl -s "$BACKEND_URL/test/cache/static/2" > /dev/null
    curl -s "$BACKEND_URL/test/cache/static/3" > /dev/null
    sleep 0.5
done

# Test de cache avec différentes tailles
echo "Test de cache avec différentes tailles..."
for size in small medium large; do
    for i in {1..5}; do
        curl -s "$BACKEND_URL/test/cache/size/$size" > /dev/null
        sleep 0.2
    done
done

echo "Tests lancés !"
echo "Note : Certains fichiers de test ont été créés (*.bin)"
echo "Vous pouvez les supprimer avec : rm *.bin"
echo "Appuyez sur Ctrl+C pour arrêter les tests..."

# Attendre que l'utilisateur arrête manuellement
wait
