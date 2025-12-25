#!/usr/bin/env bash
set -euo pipefail

: "${SCHEMA_PATH:?SCHEMA_PATH must be set}"

echo "🔍 Checking identity model invariants..."

if grep -E "PRIMARY KEY.*player_id" "$SCHEMA_PATH"; then
  echo "❌ External player_id used as PRIMARY KEY"
  exit 1
fi

if grep -E "REFERENCES .*player_id" "$SCHEMA_PATH"; then
  echo "❌ Foreign key references external player_id"
  exit 1
fi

echo "✅ Identity model invariants satisfied"
