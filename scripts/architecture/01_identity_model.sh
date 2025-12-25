#!/usr/bin/env sh
set -euo pipefail

echo "🔍 Checking identity model invariants..."

SCHEMA="tmp/schema.sql"

if [ ! -f "$SCHEMA" ]; then
  echo "❌ schema.sql not found — did 00_dump_schema.sh run?"
  exit 1
fi

# External IDs must not be primary keys
if grep -E "PRIMARY KEY.*player_id" "$SCHEMA"; then
  # echo "❌ External player_id used as PRIMARY KEY"

  echo "❌ External player_id used as PRIMARY KEY in migrations" 
  exit 1
fi

# Foreign keys must not reference external IDs
if grep -E "REFERENCES .*\\(player_id\\)" "$SCHEMA"; then
  echo "❌ Foreign key references external player_id"
  exit 1
fi

echo "✅ Identity model checks passed"
