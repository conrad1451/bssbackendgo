# CHQ: Gemini AI created file

#!/usr/bin/env sh
set -e

echo "🔍 Checking authorization rules..."

# Client-provided player IDs are forbidden
if grep -R "player_id.*:=.*r\." ./handlers; then
  echo "❌ Handler accepts player_id from request"
  exit 1
fi

# Ownership checks must use internal IDs
if grep -R "WHERE.*player_id\s*=" ./handlers ./queries; then
  echo "❌ Authorization uses external player_id"
  exit 1
fi

# Require players.id usage
if ! grep -R "players.id" ./queries >/dev/null; then
  echo "⚠️ Warning: players.id not referenced in queries"
fi

echo "✅ Authorization checks passed"
