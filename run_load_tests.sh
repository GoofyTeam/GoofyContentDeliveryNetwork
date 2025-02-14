#!/bin/bash

echo "🚀 Démarrage des tests de charge du CDN..."

# Test avec wrk (test simple de performance)
echo "📊 Test de performance avec wrk..."
echo "Test sans cache (contenu dynamique):"
wrk -t12 -c400 -d30s -s tests/wrk/benchmark.lua http://localhost:8080/test/latency/random

echo -e "\nTest avec cache (contenu statique):"
wrk -t12 -c400 -d30s http://localhost:8080/test/cache/static/1

# Test avec k6 (test de charge avancé)
echo -e "\n📈 Test de charge avancé avec k6..."
k6 run tests/k6/load_test.js

echo -e "\n✅ Tests terminés ! Vérifiez Grafana pour voir les métriques détaillées."
