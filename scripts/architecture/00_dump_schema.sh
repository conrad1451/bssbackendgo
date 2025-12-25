#!/usr/bin/env sh
set -euo pipefail

echo "📦 Dumping schema from migrations..."

mkdir -p tmp

cat migrations/*.sql > tmp/schema.sql

echo "✅ Schema snapshot generated from migrations"
