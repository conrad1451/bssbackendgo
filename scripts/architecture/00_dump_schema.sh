#!/usr/bin/env bash
set -euo pipefail

echo "📦 Dumping schema from migrations..."

BRANCH_NAME="$(git rev-parse --abbrev-ref HEAD)"
SNAPSHOT_DIR="schema/snapshots"
SNAPSHOT_PATH="${SNAPSHOT_DIR}/${BRANCH_NAME}.sql"

mkdir -p tmp "$SNAPSHOT_DIR"

cat migrations/*.sql > tmp/schema.sql
cp tmp/schema.sql "$SNAPSHOT_PATH"

echo "✅ Schema snapshot written to ${SNAPSHOT_PATH}"
