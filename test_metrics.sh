#!/bin/bash

# Durée maximale du test en secondes (peut être modifiée via TEST_DURATION=X ./test_metrics.sh)
MAX_DURATION=${TEST_DURATION:-30}
START_TIME=$(date +%s)

echo "Démarrage des tests de charge (durée maximale: ${MAX_DURATION} secondes)..."

# Fonction pour arrêter proprement tous les processus enfants
cleanup() {
    echo "Arrêt des tests..."
    pkill -P $$
    exit 0
}

# Capture du CTRL+C et autres signaux pour un arrêt propre
trap cleanup SIGINT SIGTERM

# Test de cache intensif (1000 opérations simultanées)
echo "Test de cache intensif..."
curl -s http://localhost:8080/test/cache/hammer > /dev/null &

# Test de latence aléatoire
echo "Test de latence aléatoire..."
for i in {1..20}; do
    curl -s http://localhost:8080/test/latency/random > /dev/null &
done

# Test de charge CPU sur tous les cœurs
CORES=$(nproc)
echo "Test de charge CPU sur $CORES coeurs..."
curl -s "http://localhost:8080/test/cpu/stress/$CORES" > /dev/null &

# Test de mémoire (2GB)
echo "Test de charge mémoire..."
curl -s "http://localhost:8080/test/memory/stress/2" > /dev/null &

# Test de chaos complet
echo "Démarrage du test de chaos..."
curl -s http://localhost:8080/test/mixed/chaos > /dev/null &

# Boucle pour générer du trafic constant avec vérification du temps
echo "Génération de trafic constant..."
for i in {1..5}; do
    (
        while true; do
            CURRENT_TIME=$(date +%s)
            ELAPSED_TIME=$((CURRENT_TIME - START_TIME))
            
            # Vérifier si on a dépassé la durée maximale
            if [ $ELAPSED_TIME -ge $MAX_DURATION ]; then
                echo "Durée maximale atteinte ($MAX_DURATION secondes)"
                cleanup
            fi
            
            # Mélange de requêtes réussies et d'erreurs
            curl -s http://localhost:8080/test/latency/random > /dev/null
            curl -s http://localhost:8080/test/cache/hammer > /dev/null
            sleep 0.5
        done
    ) &
done

echo "Tests de charge lancés ! Les métriques devraient apparaître dans Grafana."
echo "Le test s'arrêtera automatiquement après ${MAX_DURATION} secondes."
echo "Vous pouvez aussi appuyer sur Ctrl+C pour arrêter les tests plus tôt..."

# Attendre la durée maximale
sleep $MAX_DURATION
cleanup
