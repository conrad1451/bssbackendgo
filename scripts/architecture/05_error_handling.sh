# CHQ: Claude AI fixed grep to exclude scripts directory

#!/usr/bin/env sh
set -e

echo "🔍 Checking error handling..."

# http.Error is forbidden
if grep -R "http.Error" . --exclude-dir=scripts --exclude-dir=vendor --include="*.go"; then
  echo "❌ http.Error used — must return JSON errors"
  exit 1
fi

# Ensure writeJSONError exists
if ! grep -R "writeJSONError" . --include="*.go" >/dev/null; then
  echo "❌ writeJSONError helper not found"
  exit 1
fi

echo "✅ Error handling checks passed"
