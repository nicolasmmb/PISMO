#!/bin/bash
set -e

echo "🚀 Starting Pismo Stack..."
docker-compose up -d --build

echo ""
echo "⏳ Waiting for services..."
sleep 5

echo ""
echo "✅ Stack Ready!"
echo ""
echo "╔════════════════════════════════════════════════════════════╗"
echo "║                    🔗 QUICK ACCESS                         ║"
echo "╠════════════════════════════════════════════════════════════╣"
echo "║  API            │  http://localhost:8080                   ║"
echo "║  Health Check   │  http://localhost:8080/healthz           ║"
echo "║  Metrics        │  http://localhost:8080/metrics           ║"
echo "╠════════════════════════════════════════════════════════════╣"
echo "║  Grafana        │  http://localhost:3000/dashboards        ║"
echo "║  Prometheus     │  http://localhost:9090                   ║"
echo "║  pgAdmin        │  http://localhost:5050                   ║"
echo "╠════════════════════════════════════════════════════════════╣"
echo "║  pgAdmin Login  │  admin@pismo.com / admin                 ║"
echo "║  PostgreSQL     │  pismo / pismo (host: postgres)          ║"
echo "╚════════════════════════════════════════════════════════════╝"
echo ""
echo "📝 Test the API:"
echo "   curl http://localhost:8080/healthz"
echo ""
