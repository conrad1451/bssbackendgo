# CHQ: Gemini AI created file

#!/usr/bin/env sh
set -e

echo "🔍 Checking identity model invariants..."

# Fail if Descope IDs are used as primary keys
if grep -R "PRIMARY KEY.*player_id" .; then
  echo "❌ External player_id used as PRIMARY KEY"
  exit 1
fi

# Fail if foreign keys reference external IDs
if grep -R "REFERENCES .*player_id" .; then
  echo "❌ Foreign key references external player_id"
  exit 1
fi

# Ensure players.id exists
if ! grep -R "players.*id.*SERIAL\|players.*id.*INTEGER" . >/dev/null; then
  echo "❌ players.id primary key not found"
  exit 1
fi

echo "✅ Identity model checks passed"
