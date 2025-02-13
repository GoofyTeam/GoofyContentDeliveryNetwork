import { check, sleep } from "k6";
import http from "k6/http";
import { Rate } from "k6/metrics";

// Métriques personnalisées
const errorRate = new Rate("errors");

// Configuration des scénarios
export const options = {
  scenarios: {
    // Test de montée en charge progressive
    ramp_up: {
      executor: "ramping-vus",
      startVUs: 0,
      stages: [
        { duration: "1m", target: 20 }, // Montée progressive à 20 VUs
        { duration: "2m", target: 20 }, // Maintien à 20 VUs
        { duration: "1m", target: 0 }, // Retour à 0
      ],
    },
    // Test de pic de charge
    spike_test: {
      executor: "ramping-vus",
      startTime: "5m", // Commence après le premier test
      startVUs: 0,
      stages: [
        { duration: "10s", target: 50 }, // Pic plus modéré
        { duration: "30s", target: 50 }, // Maintien du pic
        { duration: "20s", target: 0 }, // Retour à 0
      ],
    },
  },
  thresholds: {
    http_req_duration: ["p(95)<2000"], // 95% des requêtes doivent être sous 2s
    errors: ["rate<0.1"], // Moins de 10% d'erreurs
  },
};

// URLs à tester (utilisation du CDN)
const urls = {
  static: "http://localhost:8080/test/cache/static/",
  latency: "http://localhost:8080/test/latency/random",
  download: "http://localhost:8080/test/download/small",
  compression: "http://localhost:8080/test/compression",
};

export default function () {
  // Ajout d'un délai aléatoire entre les requêtes
  sleep(Math.random() * 1);

  // Test avec cache (contenu statique)
  const staticId = Math.floor(Math.random() * 1000);
  const staticRes = http.get(`${urls.static}${staticId}`);
  check(staticRes, {
    "static-status": (r) => r.status === 200,
  });
  errorRate.add(staticRes.status !== 200);
  sleep(1);

  // Test de latence
  const latencyRes = http.get(urls.latency);
  check(latencyRes, {
    "latency-status": (r) => r.status === 200,
  });
  errorRate.add(latencyRes.status !== 200);
  sleep(1);

  // Test de téléchargement
  const downloadRes = http.get(urls.download);
  check(downloadRes, {
    "download-status": (r) => r.status === 200,
  });
  errorRate.add(downloadRes.status !== 200);
  sleep(1);

  // Test de compression
  const compressionRes = http.get(urls.compression);
  check(compressionRes, {
    "compression-status": (r) => r.status === 200,
    "compression-encoding": (r) => r.headers["Content-Encoding"] === "gzip",
  });
  errorRate.add(compressionRes.status !== 200);
  sleep(1);
}
