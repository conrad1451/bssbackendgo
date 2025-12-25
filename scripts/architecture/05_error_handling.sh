# CHQ: Gemini AI created file

#!/usr/bin/env sh
set -e

echo "🔍 Checking error handling..."

# http.Error is forbidden
if grep -R "http.Error" .; then
  echo "❌ http.Error used — must return JSON errors"
  exit 1
fi

# Ensure writeJSONError exists
if ! grep -R "writeJSONError" . >/dev/null; then
  echo "❌ writeJSONError helper not found"
  exit 1
fi

echo "✅ Error handling checks passed"
